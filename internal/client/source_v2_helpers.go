package client

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/oapi-codegen/runtime/types"
	"gopkg.in/validator.v2"
)

// Needed to create a PublicCreateSourceV2JSONBody, as union is a private field we cannot set it
// outside of this package without this helper.

func (b *PublicCreateSourceV2JSONBody) FromAny(v any) error {
	// We marshal the DTO to JSON from any object since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b.union = buf
	return nil
}

// This method is added so that NewPublicCreateSourceV2Request works
// Otherwise, the `json.Marshal(body)` call returns an empty body.
// The MarshalJSON method is usually implemented by the generator
// (see PublicCreateSourceDto_Parameters for example), but we have to create it ourselves
// for PublicCreateSourceV2JSONRequestBody because PublicCreateSourceV2Dto is a polymorphic object
// directly returned by the API and the oapi-codegen generator doesn't handle this case well.

func (t PublicCreateSourceV2JSONRequestBody) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

// Needed to create a PublicEditSourceV2JSONBody, as union is a private field we cannot set it
// outside of this package without this helper.

func (b *PublicEditSourceV2JSONBody) FromAny(v any) error {
	// We marshal the DTO to JSON from any object since oapi-codegen doesn't generate helper methods
	// for converting DTOs to request bodies when dealing with polymorphic API responses.
	buf, err := json.Marshal(v)
	if err != nil {
		return err
	}
	b.union = buf
	return nil
}

// This method is added so that NewPublicEditSourceV2Request works
// Otherwise, the call to `json.Marshal(body)` returns an empty body.

func (t PublicEditSourceV2JSONRequestBody) MarshalJSON() ([]byte, error) {
	b, err := t.union.MarshalJSON()
	return b, err
}

// Interface for parameters that are common to
// all DTOs returned by the PublicGetSourceV2 route.
type PublicGetSourceV2 interface {
	GetType() string
	GetId() types.UUID
	GetName() string
}

// ID

func (obj PublicGetAirflowSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetAthenaSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetBigQuerySourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetDatabricksSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetDbtCloudSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetDbtSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetFivetranSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetLookerSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetMicrostrategySourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetMssqlSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetMysqlSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetOracleSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetPostgresqlSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetPowerBiSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetQlikSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetQuicksightSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetRedshiftSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetSnowflakeSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetSynapseSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

func (obj PublicGetTableauSourceV2Dto) GetId() types.UUID {
	return *obj.Id
}

// Type

func (obj PublicGetAirflowSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetAthenaSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetBigQuerySourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetDatabricksSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetDbtCloudSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetDbtSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetFivetranSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetLookerSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetMicrostrategySourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetMssqlSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetMysqlSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetOracleSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetPostgresqlSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetPowerBiSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetQlikSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetQuicksightSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetRedshiftSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetSnowflakeSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetSynapseSourceV2Dto) GetType() string {
	return string(obj.Type)
}

func (obj PublicGetTableauSourceV2Dto) GetType() string {
	return string(obj.Type)
}

// Name

func (obj PublicGetAirflowSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetAthenaSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetBigQuerySourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetDatabricksSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetDbtCloudSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetDbtSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetFivetranSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetLookerSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetMicrostrategySourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetMssqlSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetMysqlSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetOracleSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetPostgresqlSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetPowerBiSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetQlikSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetQuicksightSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetRedshiftSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetSnowflakeSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetSynapseSourceV2Dto) GetName() string {
	return obj.Name
}

func (obj PublicGetTableauSourceV2Dto) GetName() string {
	return obj.Name
}

// SiffletPublicGetSourceV2Dto reprensents the `oneOf` structure of the PublicGetSourceV2Dto object.
type SiffletPublicGetSourceV2Dto struct {
	PublicGetAirflowSourceV2Dto       *PublicGetAirflowSourceV2Dto
	PublicGetAthenaSourceV2Dto        *PublicGetAthenaSourceV2Dto
	PublicGetBigQuerySourceV2Dto      *PublicGetBigQuerySourceV2Dto
	PublicGetDatabricksSourceV2Dto    *PublicGetDatabricksSourceV2Dto
	PublicGetDbtCloudSourceV2Dto      *PublicGetDbtCloudSourceV2Dto
	PublicGetDbtSourceV2Dto           *PublicGetDbtSourceV2Dto
	PublicGetFivetranSourceV2Dto      *PublicGetFivetranSourceV2Dto
	PublicGetLookerSourceV2Dto        *PublicGetLookerSourceV2Dto
	PublicGetMicrostrategySourceV2Dto *PublicGetMicrostrategySourceV2Dto
	PublicGetMssqlSourceV2Dto         *PublicGetMssqlSourceV2Dto
	PublicGetMysqlSourceV2Dto         *PublicGetMysqlSourceV2Dto
	PublicGetOracleSourceV2Dto        *PublicGetOracleSourceV2Dto
	PublicGetPostgresqlSourceV2Dto    *PublicGetPostgresqlSourceV2Dto
	PublicGetPowerBiSourceV2Dto       *PublicGetPowerBiSourceV2Dto
	PublicGetQlikSourceV2Dto          *PublicGetQlikSourceV2Dto
	PublicGetQuicksightSourceV2Dto    *PublicGetQuicksightSourceV2Dto
	PublicGetRedshiftSourceV2Dto      *PublicGetRedshiftSourceV2Dto
	PublicGetSnowflakeSourceV2Dto     *PublicGetSnowflakeSourceV2Dto
	PublicGetSynapseSourceV2Dto       *PublicGetSynapseSourceV2Dto
	PublicGetTableauSourceV2Dto       *PublicGetTableauSourceV2Dto
}

// Unmarshal JSON data into one and only one of the pointers in the struct.
func (dst *SiffletPublicGetSourceV2Dto) UnmarshalJSON(data []byte) error {
	var err error
	match := 0
	// try to unmarshal data into PublicGetAirflowSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetAirflowSourceV2Dto)
	if err == nil {
		jsonPublicGetAirflowSourceV2Dto, _ := json.Marshal(dst.PublicGetAirflowSourceV2Dto)
		if string(jsonPublicGetAirflowSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetAirflowSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetAirflowSourceV2Dto); err != nil {
				dst.PublicGetAirflowSourceV2Dto = nil
			} else if dst.PublicGetAirflowSourceV2Dto.Type != PublicGetAirflowSourceV2DtoTypeAIRFLOW {
				dst.PublicGetAirflowSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetAirflowSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetAthenaSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetAthenaSourceV2Dto)
	if err == nil {
		jsonPublicGetAthenaSourceV2Dto, _ := json.Marshal(dst.PublicGetAthenaSourceV2Dto)
		if string(jsonPublicGetAthenaSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetAthenaSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetAthenaSourceV2Dto); err != nil {
				dst.PublicGetAthenaSourceV2Dto = nil
			} else if dst.PublicGetAthenaSourceV2Dto.Type != PublicGetAthenaSourceV2DtoTypeATHENA {
				dst.PublicGetAthenaSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetAthenaSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetBigQuerySourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetBigQuerySourceV2Dto)
	if err == nil {
		jsonPublicGetBigQuerySourceV2Dto, _ := json.Marshal(dst.PublicGetBigQuerySourceV2Dto)
		if string(jsonPublicGetBigQuerySourceV2Dto) == "{}" { // empty struct
			dst.PublicGetBigQuerySourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetBigQuerySourceV2Dto); err != nil {
				dst.PublicGetBigQuerySourceV2Dto = nil
			} else if dst.PublicGetBigQuerySourceV2Dto.Type != PublicGetBigQuerySourceV2DtoTypeBIGQUERY {
				dst.PublicGetBigQuerySourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetBigQuerySourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetDatabricksSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetDatabricksSourceV2Dto)
	if err == nil {
		jsonPublicGetDatabricksSourceV2Dto, _ := json.Marshal(dst.PublicGetDatabricksSourceV2Dto)
		if string(jsonPublicGetDatabricksSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetDatabricksSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetDatabricksSourceV2Dto); err != nil {
				dst.PublicGetDatabricksSourceV2Dto = nil
			} else if dst.PublicGetDatabricksSourceV2Dto.Type != PublicGetDatabricksSourceV2DtoTypeDATABRICKS {
				dst.PublicGetDatabricksSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetDatabricksSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetDbtCloudSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetDbtCloudSourceV2Dto)
	if err == nil {
		jsonPublicGetDbtCloudSourceV2Dto, _ := json.Marshal(dst.PublicGetDbtCloudSourceV2Dto)
		if string(jsonPublicGetDbtCloudSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetDbtCloudSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetDbtCloudSourceV2Dto); err != nil {
				dst.PublicGetDbtCloudSourceV2Dto = nil
			} else if dst.PublicGetDbtCloudSourceV2Dto.Type != PublicGetDbtCloudSourceV2DtoTypeDBTCLOUD {
				dst.PublicGetDbtCloudSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetDbtCloudSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetDbtSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetDbtSourceV2Dto)
	if err == nil {
		jsonPublicGetDbtSourceV2Dto, _ := json.Marshal(dst.PublicGetDbtSourceV2Dto)
		if string(jsonPublicGetDbtSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetDbtSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetDbtSourceV2Dto); err != nil {
				dst.PublicGetDbtSourceV2Dto = nil
			} else if dst.PublicGetDbtSourceV2Dto.Type != PublicGetDbtSourceV2DtoTypeDBT {
				dst.PublicGetDbtSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetDbtSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetFivetranSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetFivetranSourceV2Dto)
	if err == nil {
		jsonPublicGetFivetranSourceV2Dto, _ := json.Marshal(dst.PublicGetFivetranSourceV2Dto)
		if string(jsonPublicGetFivetranSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetFivetranSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetFivetranSourceV2Dto); err != nil {
				dst.PublicGetFivetranSourceV2Dto = nil
			} else if dst.PublicGetFivetranSourceV2Dto.Type != PublicGetFivetranSourceV2DtoTypeFIVETRAN {
				dst.PublicGetFivetranSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetFivetranSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetLookerSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetLookerSourceV2Dto)
	if err == nil {
		jsonPublicGetLookerSourceV2Dto, _ := json.Marshal(dst.PublicGetLookerSourceV2Dto)
		if string(jsonPublicGetLookerSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetLookerSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetLookerSourceV2Dto); err != nil {
				dst.PublicGetLookerSourceV2Dto = nil
			} else if dst.PublicGetLookerSourceV2Dto.Type != PublicGetLookerSourceV2DtoTypeLOOKER {
				dst.PublicGetLookerSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetLookerSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetMicrostrategySourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetMicrostrategySourceV2Dto)
	if err == nil {
		jsonPublicGetMicrostrategySourceV2Dto, _ := json.Marshal(dst.PublicGetMicrostrategySourceV2Dto)
		if string(jsonPublicGetMicrostrategySourceV2Dto) == "{}" { // empty struct
			dst.PublicGetMicrostrategySourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetMicrostrategySourceV2Dto); err != nil {
				dst.PublicGetMicrostrategySourceV2Dto = nil
			} else if dst.PublicGetMicrostrategySourceV2Dto.Type != PublicGetMicrostrategySourceV2DtoTypeMICROSTRATEGY {
				dst.PublicGetMicrostrategySourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetMicrostrategySourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetMssqlSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetMssqlSourceV2Dto)
	if err == nil {
		jsonPublicGetMssqlSourceV2Dto, _ := json.Marshal(dst.PublicGetMssqlSourceV2Dto)
		if string(jsonPublicGetMssqlSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetMssqlSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetMssqlSourceV2Dto); err != nil {
				dst.PublicGetMssqlSourceV2Dto = nil
			} else if dst.PublicGetMssqlSourceV2Dto.Type != PublicGetMssqlSourceV2DtoTypeMSSQL {
				dst.PublicGetMssqlSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetMssqlSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetMysqlSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetMysqlSourceV2Dto)
	if err == nil {
		jsonPublicGetMysqlSourceV2Dto, _ := json.Marshal(dst.PublicGetMysqlSourceV2Dto)
		if string(jsonPublicGetMysqlSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetMysqlSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetMysqlSourceV2Dto); err != nil {
				dst.PublicGetMysqlSourceV2Dto = nil
			} else if dst.PublicGetMysqlSourceV2Dto.Type != PublicGetMysqlSourceV2DtoTypeMYSQL {
				dst.PublicGetMysqlSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetMysqlSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetOracleSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetOracleSourceV2Dto)
	if err == nil {
		jsonPublicGetOracleSourceV2Dto, _ := json.Marshal(dst.PublicGetOracleSourceV2Dto)
		if string(jsonPublicGetOracleSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetOracleSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetOracleSourceV2Dto); err != nil {
				dst.PublicGetOracleSourceV2Dto = nil
			} else if dst.PublicGetOracleSourceV2Dto.Type != PublicGetOracleSourceV2DtoTypeORACLE {
				dst.PublicGetOracleSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetOracleSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetPostgresqlSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetPostgresqlSourceV2Dto)
	if err == nil {
		jsonPublicGetPostgresqlSourceV2Dto, _ := json.Marshal(dst.PublicGetPostgresqlSourceV2Dto)
		if string(jsonPublicGetPostgresqlSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetPostgresqlSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetPostgresqlSourceV2Dto); err != nil {
				dst.PublicGetPostgresqlSourceV2Dto = nil
			} else if dst.PublicGetPostgresqlSourceV2Dto.Type != PublicGetPostgresqlSourceV2DtoTypePOSTGRESQL {
				dst.PublicGetPostgresqlSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetPostgresqlSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetPowerBiSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetPowerBiSourceV2Dto)
	if err == nil {
		jsonPublicGetPowerBiSourceV2Dto, _ := json.Marshal(dst.PublicGetPowerBiSourceV2Dto)
		if string(jsonPublicGetPowerBiSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetPowerBiSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetPowerBiSourceV2Dto); err != nil {
				dst.PublicGetPowerBiSourceV2Dto = nil
			} else if dst.PublicGetPowerBiSourceV2Dto.Type != PublicGetPowerBiSourceV2DtoTypePOWERBI {
				dst.PublicGetPowerBiSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetPowerBiSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetQlikSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetQlikSourceV2Dto)
	if err == nil {
		jsonPublicGetQlikSourceV2Dto, _ := json.Marshal(dst.PublicGetQlikSourceV2Dto)
		if string(jsonPublicGetQlikSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetQlikSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetQlikSourceV2Dto); err != nil {
				dst.PublicGetQlikSourceV2Dto = nil
			} else if dst.PublicGetQlikSourceV2Dto.Type != PublicGetQlikSourceV2DtoTypeQLIK {
				dst.PublicGetQlikSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetQlikSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetQuicksightSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetQuicksightSourceV2Dto)
	if err == nil {
		jsonPublicGetQuicksightSourceV2Dto, _ := json.Marshal(dst.PublicGetQuicksightSourceV2Dto)
		if string(jsonPublicGetQuicksightSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetQuicksightSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetQuicksightSourceV2Dto); err != nil {
				dst.PublicGetQuicksightSourceV2Dto = nil
			} else if dst.PublicGetQuicksightSourceV2Dto.Type != PublicGetQuicksightSourceV2DtoTypeQUICKSIGHT {
				dst.PublicGetQuicksightSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetQuicksightSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetRedshiftSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetRedshiftSourceV2Dto)
	if err == nil {
		jsonPublicGetRedshiftSourceV2Dto, _ := json.Marshal(dst.PublicGetRedshiftSourceV2Dto)
		if string(jsonPublicGetRedshiftSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetRedshiftSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetRedshiftSourceV2Dto); err != nil {
				dst.PublicGetRedshiftSourceV2Dto = nil
			} else if dst.PublicGetRedshiftSourceV2Dto.Type != PublicGetRedshiftSourceV2DtoTypeREDSHIFT {
				dst.PublicGetRedshiftSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetRedshiftSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetSnowflakeSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetSnowflakeSourceV2Dto)
	if err == nil {
		jsonPublicGetSnowflakeSourceV2Dto, _ := json.Marshal(dst.PublicGetSnowflakeSourceV2Dto)
		if string(jsonPublicGetSnowflakeSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetSnowflakeSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetSnowflakeSourceV2Dto); err != nil {
				dst.PublicGetSnowflakeSourceV2Dto = nil
			} else if dst.PublicGetSnowflakeSourceV2Dto.Type != PublicGetSnowflakeSourceV2DtoTypeSNOWFLAKE {
				dst.PublicGetSnowflakeSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetSnowflakeSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetSynapseSourceV2Dto
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetSynapseSourceV2Dto)
	if err == nil {
		jsonPublicGetSynapseSourceV2Dto, _ := json.Marshal(dst.PublicGetSynapseSourceV2Dto)
		if string(jsonPublicGetSynapseSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetSynapseSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetSynapseSourceV2Dto); err != nil {
				dst.PublicGetSynapseSourceV2Dto = nil
			} else if dst.PublicGetSynapseSourceV2Dto.Type != PublicGetSynapseSourceV2DtoTypeSYNAPSE {
				dst.PublicGetSynapseSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetSynapseSourceV2Dto = nil
	}

	// try to unmarshal data into PublicGetTableauSourceV2Dto.
	err = json.NewDecoder(bytes.NewBuffer(data)).Decode(&dst.PublicGetTableauSourceV2Dto)
	if err == nil {
		jsonPublicGetTableauSourceV2Dto, _ := json.Marshal(dst.PublicGetTableauSourceV2Dto)
		if string(jsonPublicGetTableauSourceV2Dto) == "{}" { // empty struct
			dst.PublicGetTableauSourceV2Dto = nil
		} else {
			if err = validator.Validate(dst.PublicGetTableauSourceV2Dto); err != nil {
				dst.PublicGetTableauSourceV2Dto = nil
			} else if dst.PublicGetTableauSourceV2Dto.Type != PublicGetTableauSourceV2DtoTypeTABLEAU {
				dst.PublicGetTableauSourceV2Dto = nil
			} else {
				match++
			}
		}
	} else {
		dst.PublicGetTableauSourceV2Dto = nil
	}

	if match > 1 {
		// If more than 1 match, reset all the pointers to nil.
		dst.PublicGetAirflowSourceV2Dto = nil
		dst.PublicGetAthenaSourceV2Dto = nil
		dst.PublicGetBigQuerySourceV2Dto = nil
		dst.PublicGetDatabricksSourceV2Dto = nil
		dst.PublicGetDbtCloudSourceV2Dto = nil
		dst.PublicGetDbtSourceV2Dto = nil
		dst.PublicGetFivetranSourceV2Dto = nil
		dst.PublicGetLookerSourceV2Dto = nil
		dst.PublicGetMicrostrategySourceV2Dto = nil
		dst.PublicGetMssqlSourceV2Dto = nil
		dst.PublicGetMysqlSourceV2Dto = nil
		dst.PublicGetOracleSourceV2Dto = nil
		dst.PublicGetPostgresqlSourceV2Dto = nil
		dst.PublicGetPowerBiSourceV2Dto = nil
		dst.PublicGetQlikSourceV2Dto = nil
		dst.PublicGetQuicksightSourceV2Dto = nil
		dst.PublicGetRedshiftSourceV2Dto = nil
		dst.PublicGetSnowflakeSourceV2Dto = nil
		dst.PublicGetSynapseSourceV2Dto = nil
		dst.PublicGetTableauSourceV2Dto = nil

		return fmt.Errorf("data matches more than one schema in oneOf(PublicCreateSourceV2201Response)")
	} else if match == 1 {
		return nil
	} else {
		return fmt.Errorf("data failed to match schemas in oneOf(PublicCreateSourceV2201Response)")
	}
}

// Gets the specific source dto from the SiffletPublicGetSourceV2Dto, that implements the PublicGetSourceV2 interface.
func (obj SiffletPublicGetSourceV2Dto) GetSourceDto() (PublicGetSourceV2, error) {
	if obj.PublicGetAirflowSourceV2Dto != nil {
		return *obj.PublicGetAirflowSourceV2Dto, nil
	}

	if obj.PublicGetAthenaSourceV2Dto != nil {
		return *obj.PublicGetAthenaSourceV2Dto, nil
	}

	if obj.PublicGetBigQuerySourceV2Dto != nil {
		return *obj.PublicGetBigQuerySourceV2Dto, nil
	}

	if obj.PublicGetDatabricksSourceV2Dto != nil {
		return *obj.PublicGetDatabricksSourceV2Dto, nil
	}

	if obj.PublicGetDbtCloudSourceV2Dto != nil {
		return *obj.PublicGetDbtCloudSourceV2Dto, nil
	}

	if obj.PublicGetDbtSourceV2Dto != nil {
		return *obj.PublicGetDbtSourceV2Dto, nil
	}

	if obj.PublicGetFivetranSourceV2Dto != nil {
		return *obj.PublicGetFivetranSourceV2Dto, nil
	}

	if obj.PublicGetLookerSourceV2Dto != nil {
		return *obj.PublicGetLookerSourceV2Dto, nil
	}

	if obj.PublicGetMicrostrategySourceV2Dto != nil {
		return *obj.PublicGetMicrostrategySourceV2Dto, nil
	}

	if obj.PublicGetMssqlSourceV2Dto != nil {
		return *obj.PublicGetMssqlSourceV2Dto, nil
	}

	if obj.PublicGetMysqlSourceV2Dto != nil {
		return *obj.PublicGetMysqlSourceV2Dto, nil
	}

	if obj.PublicGetOracleSourceV2Dto != nil {
		return *obj.PublicGetOracleSourceV2Dto, nil
	}

	if obj.PublicGetPostgresqlSourceV2Dto != nil {
		return *obj.PublicGetPostgresqlSourceV2Dto, nil
	}

	if obj.PublicGetPowerBiSourceV2Dto != nil {
		return *obj.PublicGetPowerBiSourceV2Dto, nil
	}

	if obj.PublicGetQlikSourceV2Dto != nil {
		return *obj.PublicGetQlikSourceV2Dto, nil
	}

	if obj.PublicGetQuicksightSourceV2Dto != nil {
		return *obj.PublicGetQuicksightSourceV2Dto, nil
	}

	if obj.PublicGetRedshiftSourceV2Dto != nil {
		return *obj.PublicGetRedshiftSourceV2Dto, nil
	}

	if obj.PublicGetSnowflakeSourceV2Dto != nil {
		return *obj.PublicGetSnowflakeSourceV2Dto, nil
	}

	if obj.PublicGetSynapseSourceV2Dto != nil {
		return *obj.PublicGetSynapseSourceV2Dto, nil
	}

	if obj.PublicGetTableauSourceV2Dto != nil {
		return *obj.PublicGetTableauSourceV2Dto, nil
	}

	return nil, fmt.Errorf("data failed to match schemas in oneOf(PublicCreateSourceV2201Response)")
}

// Creates a SiffletPublicGetSourceV2Dto from a PublicPageDtoPublicGetSourceV2Dto_Data_Item.
func (dst *SiffletPublicGetSourceV2Dto) FromPublicPageDtoPublicGetSourceV2DtoDataItem(dto PublicPageDtoPublicGetSourceV2Dto_Data_Item) error {
	return dst.UnmarshalJSON(dto.union)
}

type PublicGetSourceV2DataItem struct {
	Name *string
	Id   *string
}

func (dst *PublicGetSourceV2DataItem) FromPublicPageDtoPublicGetSourceV2DtoDataItem(dto PublicPageDtoPublicGetSourceV2Dto_Data_Item) error {
	m := make(map[string]any)
	if err := json.Unmarshal(dto.union, &m); err != nil {
		return fmt.Errorf("couldn't parse parameters as JSON: %s", err)
	}

	name, ok := m["name"]
	if !ok {
		return fmt.Errorf("'name' field not found in parameters")
	}
	nameStr, ok := name.(string)
	if !ok {
		return fmt.Errorf("expected 'name' field to be a string, got %T", name)
	}

	id, ok := m["id"]
	if !ok {
		return fmt.Errorf("'id' field not found in parameters")
	}
	idStr, ok := id.(string)
	if !ok {
		return fmt.Errorf("expected 'id' field to be a string, got %T", id)
	}

	dst.Name = &nameStr
	dst.Id = &idStr
	return nil
}
