// SPDX-License-Identifier: Apache-2.0
// Copyright (C) 2026 SukramJ.

package fabricui

import (
	"fmt"

	"github.com/SukramJ/go-fabric/schema"
)

// deviceTypeLabels are the operator-facing names of the contract's
// device_type_name that differ from the Matter device type's own name (in
// the comment). go-fabric v0.16.0 keeps one name, schema.DeviceTypeName, and
// leaves a host's labels to the host.
var deviceTypeLabels = map[uint16]string{
	0x000A: "Door Lock",               // DoorLock
	0x000F: "Generic Switch",          // GenericSwitch
	0x0015: "Contact Sensor",          // ContactSensor
	0x0027: "Mode Select",             // ModeSelect
	0x002C: "Air Quality Sensor",      // AirQualitySensor
	0x002D: "Air Purifier",            // AirPurifier
	0x0043: "Water Leak Detector",     // WaterLeakDetector
	0x0073: "Laundry Washer",          // LaundryWasher
	0x0074: "Robotic Vacuum Cleaner",  // RoboticVacuumCleaner
	0x0076: "Smoke / CO Alarm",        // SmokeCoAlarm
	0x007A: "Extractor Hood",          // ExtractorHood
	0x007C: "Laundry Dryer",           // LaundryDryer
	0x0100: "On/Off Light",            // OnOffLight
	0x0101: "Dimmable Light",          // DimmableLight
	0x0106: "Light Sensor",            // LightSensor
	0x0107: "Occupancy Sensor",        // OccupancySensor
	0x010A: "On/Off Plug-in Unit",     // OnOffPlugInUnit
	0x010C: "Color Temperature Light", // ColorTemperatureLight
	0x010D: "Extended Color Light",    // ExtendedColorLight
	0x0202: "Window Covering",         // WindowCovering
	0x0302: "Temperature Sensor",      // TemperatureSensor
	0x0305: "Pressure Sensor",         // PressureSensor
	0x0306: "Flow Sensor",             // FlowSensor
	0x0307: "Humidity Sensor",         // HumiditySensor
	0x0510: "Electrical Sensor",       // ElectricalSensor
}

// deviceTypeName is the contract's device_type_name: empty for no device
// type, the label or the schema's name, and the id in hex for a device type
// neither knows.
func deviceTypeName(id uint16) string {
	if id == 0 {
		return ""
	}
	if label, ok := deviceTypeLabels[id]; ok {
		return label
	}
	if name, ok := schema.DeviceTypeName(schema.DeviceTypeID(id)); ok {
		return name
	}
	return fmt.Sprintf("0x%04X", id)
}
