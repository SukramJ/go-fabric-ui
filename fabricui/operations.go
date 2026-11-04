// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
)

// EventLogRecord is the one event type of GET /operation/log/stream: its
// payload is a [LogRecord] as GET /operation/log serves it.
const EventLogRecord = "log_record"

// Operations is the daemon's operation as the Operation page reads and
// drives it: the level it logs at, the bounded ring of its recent records
// and their stream, and what it is. A host implements it over its logger —
// go-fabric's logring package is that implementation once it lands
// (go-fabric issue #393): a slog.Handler with a LevelVar and a ring in
// front of the host's own handler. [MemoryOperations] is the test double.
// Every method is safe for concurrent use; nothing it returns carries a
// passcode.
type Operations interface {
	// LogLevel is the level in force: a record below it is not logged.
	LogLevel() slog.Level
	// SetLogLevel changes it until the daemon restarts.
	SetLogLevel(l slog.Level) error
	// Log are the records the ring holds with a sequence number of since
	// or more, oldest first; with limit > 0 only the newest limit of them.
	Log(since uint64, limit int) []LogRecord
	// ObserveLog registers fn for every record the ring takes from now
	// on and returns the function that unregisters it. fn runs on the
	// logging goroutine and must not block.
	ObserveLog(fn func(LogRecord)) (unsubscribe func())
	// Info is the daemon's version and uptime.
	Info() OperationInfo
}

// LogRecord is one record of the log ring.
type LogRecord struct {
	// Seq increases by one per record the ring takes, from 1.
	Seq     uint64
	Time    time.Time
	Level   slog.Level
	Message string
	// Attrs are the record's attributes as text (slog.Value.String), a
	// group's under dotted keys.
	Attrs map[string]string
}

// OperationInfo is what the daemon is.
type OperationInfo struct {
	// Version is the daemon's version, as its host names it.
	Version string
	// GoFabric is the go-fabric module version it is built with
	// ([GoFabricVersion]); empty when the build does not say.
	GoFabric string
	// StartedAt is when the daemon started; Uptime the time since.
	StartedAt time.Time
	Uptime    time.Duration
}

// GoFabricVersion is the version of the go-fabric module in the running
// binary's build information; empty when the binary carries none (a test
// binary built outside module mode).
func GoFabricVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		return ""
	}
	for _, d := range bi.Deps {
		if d.Path == "github.com/SukramJ/go-fabric" {
			if d.Replace != nil {
				return d.Replace.Version
			}
			return d.Version
		}
	}
	return ""
}

// logLevels are the level names the contract carries (LogLevel), lowest
// first: slog's four, lower-case.
var logLevels = []struct {
	name  string
	level slog.Level
}{
	{"debug", slog.LevelDebug},
	{"info", slog.LevelInfo},
	{"warn", slog.LevelWarn},
	{"error", slog.LevelError},
}

// levelName is l's contract name: the highest named level at or below l,
// "debug" below every name.
func levelName(l slog.Level) string {
	name := logLevels[0].name
	for _, n := range logLevels {
		if l >= n.level {
			name = n.name
		}
	}
	return name
}

// parseLevel is the level a contract name names.
func parseLevel(s string) (slog.Level, bool) {
	for _, n := range logLevels {
		if n.name == s {
			return n.level, true
		}
	}
	return 0, false
}

// --- the routes ---

type logLevelResponse struct {
	Level  string   `json:"level"`
	Levels []string `json:"levels"`
}

type logLevelRequest struct {
	Level string `json:"level"`
}

type logRecordResponse struct {
	Seq   uint64            `json:"seq"`
	Time  string            `json:"time"`
	Level string            `json:"level"`
	Msg   string            `json:"msg"`
	Attrs map[string]string `json:"attrs"`
}

type logRecordList struct {
	Records []logRecordResponse `json:"records"`
}

type operationInfoResponse struct {
	Version   string `json:"version"`
	GoFabric  string `json:"go_fabric"`
	UptimeS   int64  `json:"uptime_s"`
	StartedAt string `json:"started_at"`
}

// The window of GET /operation/log (contract LogSince, LogLimit).
const (
	defaultLogLimit = 200
	maxLogLimit     = 2000
)

func logRecordOf(r *LogRecord) logRecordResponse {
	attrs := r.Attrs
	if attrs == nil {
		attrs = map[string]string{}
	}
	return logRecordResponse{
		Seq: r.Seq, Time: r.Time.UTC().Format(time.RFC3339Nano), Level: levelName(r.Level), Msg: r.Message, Attrs: attrs,
	}
}

func (h *handler) logLevelBody() logLevelResponse {
	out := logLevelResponse{Level: levelName(h.ops.LogLevel()), Levels: make([]string, 0, len(logLevels))}
	for _, n := range logLevels {
		out.Levels = append(out.Levels, n.name)
	}
	return out
}

// opsUnready answers an operation route while the host attached no port.
func opsUnready(w http.ResponseWriter, r *http.Request) {
	unready(w, r, "Operation not available", "the host serves no operations port (Options.Operations is nil)")
}

// operationInfo serves GET /operation/info.
func (h *handler) operationInfo(w http.ResponseWriter, r *http.Request) {
	if h.ops == nil {
		opsUnready(w, r)
		return
	}
	in := h.ops.Info()
	writeJSON(w, operationInfoResponse{
		Version: in.Version, GoFabric: in.GoFabric, UptimeS: int64(in.Uptime / time.Second),
		StartedAt: in.StartedAt.UTC().Format(time.RFC3339),
	})
}

// logLevel serves GET /operation/log/level.
func (h *handler) logLevel(w http.ResponseWriter, r *http.Request) {
	if h.ops == nil {
		opsUnready(w, r)
		return
	}
	writeJSON(w, h.logLevelBody())
}

// setLogLevel serves PUT /operation/log/level (admin).
func (h *handler) setLogLevel(w http.ResponseWriter, r *http.Request) {
	if h.ops == nil {
		opsUnready(w, r)
		return
	}
	var body logLevelRequest
	if err := decodeJSON(w, r, &body); err != nil {
		badBody(w, r, err)
		return
	}
	l, ok := parseLevel(body.Level)
	if !ok {
		writeProblem(w, r, http.StatusBadRequest, codeBadRequest, "Invalid log level",
			"level "+strconv.Quote(body.Level)+" is not one of debug, info, warn, error")
		return
	}
	if err := h.ops.SetLogLevel(l); err != nil {
		h.serverError(w, r, "Setting the log level failed", err)
		return
	}
	h.log.InfoContext(r.Context(), "fabricui.operation.log_level", "level", levelName(l))
	writeJSON(w, h.logLevelBody())
}

// logRecords serves GET /operation/log.
func (h *handler) logRecords(w http.ResponseWriter, r *http.Request) {
	if h.ops == nil {
		opsUnready(w, r)
		return
	}
	since, limit, err := logWindow(r)
	if err != nil {
		h.writeAPIError(w, r, "Reading the log failed", err)
		return
	}
	recs := h.ops.Log(since, limit)
	out := logRecordList{Records: make([]logRecordResponse, 0, len(recs))}
	for i := range recs {
		out.Records = append(out.Records, logRecordOf(&recs[i]))
	}
	writeJSON(w, out)
}

// logWindow is the since and limit of GET /operation/log.
func logWindow(r *http.Request) (since uint64, limit int, err error) {
	sinceQ, err := queryInt(r, "since")
	if err != nil {
		return 0, 0, err
	}
	limitQ, err := queryInt(r, "limit")
	if err != nil {
		return 0, 0, err
	}
	if sinceQ != nil {
		if *sinceQ < 0 {
			return 0, 0, badRequest("since: %d is not a sequence number", *sinceQ)
		}
		since = uint64(*sinceQ)
	}
	limit = defaultLogLimit
	if limitQ != nil {
		if *limitQ < 1 || *limitQ > maxLogLimit {
			return 0, 0, badRequest("limit: %d is not within 1..%d", *limitQ, maxLogLimit)
		}
		limit = *limitQ
	}
	return since, limit, nil
}

// logStream serves GET /operation/log/stream: every record the ring takes
// while the stream is open, with the stream machinery of the node's
// (buffer, slow-client drop, heartbeat).
func (h *handler) logStream(w http.ResponseWriter, r *http.Request) {
	if h.ops == nil {
		opsUnready(w, r)
		return
	}
	h.serveSSE(r.Context(), w, func(fn func(Event)) func() {
		return h.ops.ObserveLog(func(rec LogRecord) {
			fn(Event{Type: EventLogRecord, Data: logRecordOf(&rec)})
		})
	})
}

// --- the test double ---

// DefaultMemoryLogCapacity is the ring size of a [MemoryOperations] built
// with a capacity below one.
const DefaultMemoryLogCapacity = 1000

// MemoryOperations is an in-memory [Operations]: a ring of the newest
// records, a level and a fixed [OperationInfo]. A test or a demo appends
// records with [MemoryOperations.Record]; a record below the level is
// dropped as a logger would drop it. It is safe for concurrent use.
type MemoryOperations struct {
	// Now is the clock Info's uptime runs on; nil is time.Now.
	Now func() time.Time

	mu       sync.Mutex
	level    slog.Level
	capacity int
	ring     []LogRecord
	seq      uint64
	info     OperationInfo
	nextObs  int
	obs      map[int]func(LogRecord)
}

var _ Operations = (*MemoryOperations)(nil)

// NewMemoryOperations returns a ring of capacity records (below one is
// [DefaultMemoryLogCapacity]) at level info, reporting info (its Uptime
// is computed from StartedAt).
func NewMemoryOperations(capacity int, info OperationInfo) *MemoryOperations {
	if capacity < 1 {
		capacity = DefaultMemoryLogCapacity
	}
	return &MemoryOperations{level: slog.LevelInfo, capacity: capacity, info: info, obs: map[int]func(LogRecord){}}
}

// Record appends a record at t unless l is below the level, and hands it
// to the observers; it returns the record and whether it was taken.
func (m *MemoryOperations) Record(t time.Time, l slog.Level, msg string, attrs map[string]string) (LogRecord, bool) {
	m.mu.Lock()
	if l < m.level {
		m.mu.Unlock()
		return LogRecord{}, false
	}
	m.seq++
	rec := LogRecord{Seq: m.seq, Time: t, Level: l, Message: msg, Attrs: copyAttrs(attrs)}
	m.ring = append(m.ring, rec)
	if len(m.ring) > m.capacity {
		m.ring = append([]LogRecord(nil), m.ring[len(m.ring)-m.capacity:]...)
	}
	fns := make([]func(LogRecord), 0, len(m.obs))
	for _, fn := range m.obs {
		fns = append(fns, fn)
	}
	m.mu.Unlock()
	for _, fn := range fns {
		fn(rec)
	}
	return rec, true
}

func copyAttrs(in map[string]string) map[string]string {
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

// LogLevel implements [Operations].
func (m *MemoryOperations) LogLevel() slog.Level {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.level
}

// SetLogLevel implements [Operations].
func (m *MemoryOperations) SetLogLevel(l slog.Level) error {
	m.mu.Lock()
	m.level = l
	m.mu.Unlock()
	return nil
}

// Log implements [Operations].
func (m *MemoryOperations) Log(since uint64, limit int) []LogRecord {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []LogRecord{}
	for i := range m.ring {
		if m.ring[i].Seq >= since {
			r := m.ring[i]
			r.Attrs = copyAttrs(r.Attrs)
			out = append(out, r)
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[len(out)-limit:]
	}
	return out
}

// ObserveLog implements [Operations].
func (m *MemoryOperations) ObserveLog(fn func(LogRecord)) (unsubscribe func()) {
	m.mu.Lock()
	id := m.nextObs
	m.nextObs++
	m.obs[id] = fn
	m.mu.Unlock()
	return func() {
		m.mu.Lock()
		delete(m.obs, id)
		m.mu.Unlock()
	}
}

// Info implements [Operations].
func (m *MemoryOperations) Info() OperationInfo {
	now := time.Now
	if m.Now != nil {
		now = m.Now
	}
	m.mu.Lock()
	in := m.info
	m.mu.Unlock()
	if !in.StartedAt.IsZero() {
		in.Uptime = max(now().Sub(in.StartedAt), 0)
	}
	return in
}
