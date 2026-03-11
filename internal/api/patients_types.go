package api

type PatientAddress struct {
	StreetName string `json:"streetName"`
	CoAddress  string `json:"coAddress"`
	City       string `json:"city"`
	ZipCode    string `json:"zipCode"`
}

type ListedClinic struct {
	HsaID string `json:"hsaId"`
}

type PatientOrganization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PatientDepartment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type County struct {
	Code string `json:"code"`
}

type PatientType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type Patient struct {
	ID                  string               `json:"id"`
	PersonalNumber      string               `json:"personalNumber"`
	BirthDate           string               `json:"birthDate"`
	Gender              string               `json:"gender"`
	FirstName           string               `json:"firstName"`
	LastName            string               `json:"lastName"`
	Address             PatientAddress       `json:"address"`
	Nationality         string               `json:"nationality"`
	Email               string               `json:"email"`
	MobilePhoneNumber   string               `json:"mobilePhoneNumber"`
	WorkPhoneNumber     string               `json:"workPhoneNumber"`
	HomePhoneNumber     string               `json:"homePhoneNumber"`
	ListedClinic        ListedClinic         `json:"listedClinic"`
	InterpreterNeeded   bool                 `json:"interpreterNeeded"`
	InterpreterLanguage string               `json:"interpreterLanguage"`
	Organization        *PatientOrganization `json:"organization"`
	Department          *PatientDepartment   `json:"department"`
	County              County               `json:"county"`
	PatientType         PatientType          `json:"patientType"`
}

type CreatePatientListInfo struct {
	Name       string `json:"name"`
	HsaID      string `json:"hsaId"`
	CountyName string `json:"countyName"`
}

type CreatePatientRequest struct {
	PersonalNumber string                `json:"personalNumber"`
	ListInfo       CreatePatientListInfo `json:"listInfo"`
	PatientType    string                `json:"patientType"`
	Email          string                `json:"email,omitempty"`
	MobileNumber   string                `json:"mobileNumber,omitempty"`
}

type TempAddress struct {
	TempStreetName string `json:"tempStreetName"`
	TempZipCode    string `json:"tempZipCode"`
	TempCoAddress  string `json:"tempCoAddress"`
	TempCity       string `json:"tempCity"`
	TempValidUntil string `json:"tempValidUntil"`
}

type CreatedPatientContact struct {
	StreetName  string      `json:"streetName"`
	ZipCode     string      `json:"zipCode"`
	CoAddress   string      `json:"coAddress"`
	City        string      `json:"city"`
	HomePhone   string      `json:"homePhone"`
	WorkPhone   string      `json:"workPhone"`
	Mobile      string      `json:"mobile"`
	Email       string      `json:"email"`
	TempAddress TempAddress `json:"tempAddress"`
}

type CreatedPatientListInfo struct {
	Name       string `json:"name"`
	HsaID      string `json:"hsaId"`
	CountyName string `json:"countyName"`
}

// CreatedPatient is the v1 response shape from POST /v1/patients.
// It differs from the v2 Patient type returned by GET /v2/patients.
type CreatedPatient struct {
	ID                  string                 `json:"id"`
	PersonalNumber      string                 `json:"personalNumber"`
	FirstName           string                 `json:"firstName"`
	LastName            string                 `json:"lastName"`
	BirthDate           string                 `json:"birthDate"`
	Gender              string                 `json:"gender"`
	Contact             CreatedPatientContact  `json:"contact"`
	ListInfo            CreatedPatientListInfo `json:"listInfo"`
	PatientType         string                 `json:"patientType"`
	Nationality         string                 `json:"nationality"`
	InterpreterNeed     bool                   `json:"interpreterNeed"`
	InterpreterLanguage string                 `json:"interpreterLanguage"`
	RegisterAsOf        string                 `json:"registerAsOf"`
}
