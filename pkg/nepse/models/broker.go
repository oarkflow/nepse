package models

type Broker struct {
	ActiveStatus                  string  `json:"activeStatus"`
	AuthorizedContactPerson       *string `json:"authorizedContactPerson"`
	AuthorizedContactPersonNumber string  `json:"authorizedContactPersonNumber"`
	ClearingMemberId              *string `json:"clearingMemberId"`
	DistrictList                  []struct {
		DistrictName string `json:"districtName"`
		Id           int    `json:"id"`
		ProvinceId   int    `json:"provinceId"`
		Status       string `json:"status"`
	} `json:"districtList"`
	Id                   int    `json:"id"`
	IsDealer             string `json:"isDealer"`
	MemberBranchMappings []struct {
		ActiveStatus   string  `json:"activeStatus"`
		BranchHead     *string `json:"branchHead"`
		BranchLocation string  `json:"branchLocation"`
		BranchName     string  `json:"branchName"`
		District       struct {
			DistrictName string `json:"districtName"`
			Id           int    `json:"id"`
			ProvinceId   int    `json:"provinceId"`
			Status       string `json:"status"`
		} `json:"district"`
		Id           int `json:"id"`
		Municipality struct {
			DistrictId       int    `json:"districtId"`
			Id               int    `json:"id"`
			MunicipalityName string `json:"municipalityName"`
			Status           string `json:"status"`
		} `json:"municipality"`
		PhoneNumber *string `json:"phoneNumber"`
		Province    struct {
			Description string `json:"description"`
			Id          int    `json:"id"`
			Name        string `json:"name"`
			Status      string `json:"status"`
		} `json:"province"`
	} `json:"memberBranchMappings"`
	MemberCode           int    `json:"memberCode"`
	MemberName           string `json:"memberName"`
	MemberTMSLinkMapping *struct {
		TmsLink string `json:"tmsLink"`
	} `json:"memberTMSLinkMapping"`
	MembershipTypeMaster struct {
		HibernateLazyInitializer struct {
		} `json:"hibernateLazyInitializer"`
		Id             int    `json:"id"`
		MembershipType string `json:"membershipType"`
	} `json:"membershipTypeMaster"`
	Municipalities []struct {
		DistrictId       int    `json:"districtId"`
		Id               int    `json:"id"`
		MunicipalityName string `json:"municipalityName"`
		Status           string `json:"status"`
	} `json:"municipalities"`
	ProvinceList []struct {
		Description string `json:"description"`
		Id          int    `json:"id"`
		Name        string `json:"name"`
		Status      string `json:"status"`
	} `json:"provinceList"`
}

type BrokerData struct {
	Content []Broker `json:"content"`
	Paging
}
