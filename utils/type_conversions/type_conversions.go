// SPDX-License-Identifier: Apache-2.0
// Copyright Pionix GmbH and Contributors to EVerest

package type_conversions

import (
	"fmt"
	"strings"
	"time"

	"github.com/enbility/eebus-go/api"
	uc_api "github.com/enbility/eebus-go/usecases/api"
	"github.com/enbility/eebus-grpc/rpc_services/common_types"
	"github.com/enbility/eebus-grpc/rpc_services/control_service"
	"github.com/enbility/spine-go/model"
)

// deviceTypeMap maps proto DeviceType_Enum values to the SPINE model strings.
// The proto enum uses SCREAMING_SNAKE_CASE; SPINE uses PascalCase — they do not
// match, so a manual mapping is required.
var deviceTypeMap = map[control_service.DeviceType_Enum]model.DeviceTypeType{
	control_service.DeviceType_DISHWASHER:                model.DeviceTypeTypeDishwasher,
	control_service.DeviceType_DRYER:                     model.DeviceTypeTypeDryer,
	control_service.DeviceType_ENVIRONMENT_SENSOR:        model.DeviceTypeTypeEnvironmentSensor,
	control_service.DeviceType_GENERIC:                   model.DeviceTypeTypeGeneric,
	control_service.DeviceType_HEAT_GENERATION_SYSTEM:    model.DeviceTypeTypeHeatgenerationSystem,
	control_service.DeviceType_HEAT_SINK_SYSTEM:          model.DeviceTypeTypeHeatsinkSystem,
	control_service.DeviceType_HEAT_STORAGE_SYSTEM:       model.DeviceTypeTypeHeatstorageSystem,
	control_service.DeviceType_HVAC_CONTROLLER:           model.DeviceTypeTypeHVACController,
	control_service.DeviceType_SUBMETER:                  model.DeviceTypeTypeSubmeter,
	control_service.DeviceType_WASHER:                    model.DeviceTypeTypeWasher,
	control_service.DeviceType_ELECTRICITY_SUPPLY_SYSTEM: model.DeviceTypeTypeElectricitySupplySystem,
	control_service.DeviceType_ENERGY_MANAGEMENT_SYSTEM:  model.DeviceTypeTypeEnergyManagementSystem,
	control_service.DeviceType_INVERTER:                  model.DeviceTypeTypeInverter,
	control_service.DeviceType_CHARGING_STATION:          model.DeviceTypeTypeChargingStation,
}

// entityTypeMap maps proto EntityType_Enum values to the SPINE model strings.
var entityTypeMap = map[control_service.EntityType_Enum]model.EntityTypeType{
	control_service.EntityType_Battery:                       model.EntityTypeTypeBattery,
	control_service.EntityType_Compressor:                    model.EntityTypeTypeCompressor,
	control_service.EntityType_DeviceInformation:             model.EntityTypeTypeDeviceInformation,
	control_service.EntityType_DHWCircuit:                    model.EntityTypeTypeDHWCircuit,
	control_service.EntityType_DHWStorage:                    model.EntityTypeTypeDHWStorage,
	control_service.EntityType_Dishwasher:                    model.EntityTypeTypeDishwasher,
	control_service.EntityType_Dryer:                         model.EntityTypeTypeDryer,
	control_service.EntityType_ElectricalImmersionHeater:     model.EntityTypeTypeElectricalImmersionHeater,
	control_service.EntityType_Fan:                           model.EntityTypeTypeFan,
	control_service.EntityType_GasHeatingAppliance:           model.EntityTypeTypeGasHeatingAppliance,
	control_service.EntityType_Generic:                       model.EntityTypeTypeGeneric,
	control_service.EntityType_HeatingBufferStorage:          model.EntityTypeTypeHeatingBufferStorage,
	control_service.EntityType_HeatingCircuit:                model.EntityTypeTypeHeatingCircuit,
	control_service.EntityType_HeatingObject:                 model.EntityTypeTypeHeatingObject,
	control_service.EntityType_HeatingZone:                   model.EntityTypeTypeHeatingZone,
	control_service.EntityType_HeatPumpAppliance:             model.EntityTypeTypeHeatPumpAppliance,
	control_service.EntityType_HeatSinkCircuit:               model.EntityTypeTypeHeatSinkCircuit,
	control_service.EntityType_HeatSourceCircuit:             model.EntityTypeTypeHeatSourceCircuit,
	control_service.EntityType_HeatSourceUnit:                model.EntityTypeTypeHeatSourceUnit,
	control_service.EntityType_HvacController:                model.EntityTypeTypeHvacController,
	control_service.EntityType_HvacRoom:                      model.EntityTypeTypeHvacRoom,
	control_service.EntityType_InstantDHWHeater:              model.EntityTypeTypeInstantDHWheater,
	control_service.EntityType_Inverter:                      model.EntityTypeTypeInverter,
	control_service.EntityType_OilHeatingAppliance:           model.EntityTypeTypeOilHeatingAppliance,
	control_service.EntityType_Pump:                          model.EntityTypeTypePump,
	control_service.EntityType_RefrigerantCircuit:            model.EntityTypeTypeRefrigerantCircuit,
	control_service.EntityType_SmartEnergyAppliance:          model.EntityTypeTypeSmartEnergyAppliance,
	control_service.EntityType_SolarDHWStorage:               model.EntityTypeTypeSolarDHWStorage,
	control_service.EntityType_SolarThermalCircuit:           model.EntityTypeTypeSolarThermalCircuit,
	control_service.EntityType_SubMeterElectricity:           model.EntityTypeTypeSubMeterElectricity,
	control_service.EntityType_TemperatureSensor:             model.EntityTypeTypeTemperatureSensor,
	control_service.EntityType_Washer:                        model.EntityTypeTypeWasher,
	control_service.EntityType_BatterySystem:                 model.EntityTypeTypeBatterySystem,
	control_service.EntityType_ElectricityGenerationSystem:   model.EntityTypeTypeElectricityGenerationSystem,
	control_service.EntityType_ElectricityStorageSystem:      model.EntityTypeTypeElectricityStorageSystem,
	control_service.EntityType_GridConnectionPointOfPremises: model.EntityTypeTypeGridConnectionPointOfPremises,
	control_service.EntityType_Household:                     model.EntityTypeTypeHousehold,
	control_service.EntityType_PVSystem:                      model.EntityTypeTypePVSystem,
	control_service.EntityType_EV:                            model.EntityTypeTypeEV,
	control_service.EntityType_EVSE:                          model.EntityTypeTypeEVSE,
	control_service.EntityType_ChargingOutlet:                model.EntityTypeTypeChargingOutlet,
	control_service.EntityType_CEM:                           model.EntityTypeTypeCEM,
	control_service.EntityType_PV:                            model.EntityTypeTypePV,
	control_service.EntityType_PVESHybrid:                    model.EntityTypeTypePVESHybrid,
	control_service.EntityType_ElectricalStorage:             model.EntityTypeTypeElectricalStorage,
	control_service.EntityType_PVString:                      model.EntityTypeTypePVString,
	control_service.EntityType_GridGuard:                     model.EntityTypeTypeGridGuard,
	control_service.EntityType_ControllableSystem:            model.EntityTypeTypeControllableSystem,
}

func init() {
	// Verify that every DeviceType_Enum value (except UNKNOWN) has a mapping.
	for i := range control_service.DeviceType_Enum_name {
		e := control_service.DeviceType_Enum(i)
		if e == control_service.DeviceType_UNKNOWN {
			continue
		}
		if _, ok := deviceTypeMap[e]; !ok {
			panic(fmt.Sprintf("deviceTypeMap missing entry for %s", e))
		}
	}
	// Verify that every EntityType_Enum value (except UNKNOWN) has a mapping.
	for i := range control_service.EntityType_Enum_name {
		e := control_service.EntityType_Enum(i)
		if e == control_service.EntityType_UNKNOWN {
			continue
		}
		if _, ok := entityTypeMap[e]; !ok {
			panic(fmt.Sprintf("entityTypeMap missing entry for %s", e))
		}
	}
}

// ConvertRPCDeviceType converts a proto DeviceType_Enum to the SPINE model.DeviceTypeType.
func ConvertRPCDeviceType(dt control_service.DeviceType_Enum) model.DeviceTypeType {
	if v, ok := deviceTypeMap[dt]; ok {
		return v
	}
	return model.DeviceTypeTypeGeneric
}

// ConvertRPCEntityType converts a proto EntityType_Enum to the SPINE model.EntityTypeType.
func ConvertRPCEntityType(et control_service.EntityType_Enum) model.EntityTypeType {
	if v, ok := entityTypeMap[et]; ok {
		return v
	}
	return model.EntityTypeTypeGeneric
}

func ConvertRPCEntityAddress(address *common_types.EntityAddress) []model.AddressEntityType {
	var entity_address []model.AddressEntityType
	for _, entry := range address.GetEntityAddress() {
		entity_address = append(
			entity_address,
			model.AddressEntityType(entry),
		)
	}
	return entity_address
}

func ConvertEEBUSEntityAddress(address []model.AddressEntityType) *common_types.EntityAddress {
	var entity_address []uint32
	for _, id := range address {
		entity_address = append(entity_address, uint32(id))
	}
	return &common_types.EntityAddress{
		EntityAddress: entity_address,
	}
}

func ConvertRPCLoadLimit(limit *common_types.LoadLimit) uc_api.LoadLimit {
	return uc_api.LoadLimit{
		Duration:       time.Duration(limit.GetDurationNanoseconds()) * time.Nanosecond,
		IsChangeable:   limit.GetIsChangeable(),
		IsActive:       limit.GetIsActive(),
		Value:          limit.GetValue(),
		DeleteDuration: limit.GetDeleteDuration(),
	}
}

func ConvertEEBUSLoadLimit(limit *uc_api.LoadLimit) common_types.LoadLimit {
	return common_types.LoadLimit{
		DurationNanoseconds: limit.Duration.Nanoseconds(),
		IsChangeable:        limit.IsChangeable,
		IsActive:            limit.IsActive,
		Value:               limit.Value,
		DeleteDuration:      limit.DeleteDuration,
	}
}

func ConvertEEBUSEventType(event api.EventType) control_service.UseCaseEvent {
	//split string at "-"
	parts := strings.Split(string(event), "-")
	actor := ConvertEEBUSEventTypeActor(parts[0])
	name := ConvertEEBUSEventTypeName(parts[1])
	return control_service.UseCaseEvent{
		UseCase: &control_service.UseCase{
			Actor: actor,
			Name:  name,
		},
		Event: parts[2],
	}
}

func ConvertEEBUSEventTypeActor(actor string) control_service.UseCase_ActorType_Enum {
	switch actor {
	case "cs":
		return control_service.UseCase_ActorType_ControllableSystem
	case "eg":
		return control_service.UseCase_ActorType_EnergyGuard
	case "cem":
		return control_service.UseCase_ActorType_CEM
	case "ma":
		return control_service.UseCase_ActorType_MonitoringAppliance
	default:
		return control_service.UseCase_ActorType_UNKNOWN
	}
}

func ConvertEEBUSEventTypeName(name string) control_service.UseCase_NameType_Enum {
	switch name {
	case "cevc":
		return control_service.UseCase_NameType_coordinatedEvCharging
	case "evcc":
		return control_service.UseCase_NameType_evCommissioningAndConfiguration
	case "evcem":
		return control_service.UseCase_NameType_measurementOfElectricityDuringEvCharging
	case "evsecc":
		return control_service.UseCase_NameType_evseCommissioningAndConfiguration
	case "evsoc":
		return control_service.UseCase_NameType_evStateOfCharge
	case "opev":
		return control_service.UseCase_NameType_optimizationOfSelfConsumptionDuringEvCharging
	case "vabd":
		return control_service.UseCase_NameType_visualizationOfAggregatedBatteryData
	case "vapd":
		return control_service.UseCase_NameType_visualizationOfAggregatedPhotovoltaicData
	case "lpc":
		return control_service.UseCase_NameType_limitationOfPowerConsumption
	case "lpp":
		return control_service.UseCase_NameType_limitationOfPowerProduction
	case "mgcp":
		return control_service.UseCase_NameType_monitoringOfGridConnectionPoint
	case "mpc":
		return control_service.UseCase_NameType_monitoringOfPowerConsumption
	default:
		return control_service.UseCase_NameType_UNKNOWN
	}
}
