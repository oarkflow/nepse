package nepse

import (
	_ "embed"
)

type AuthenticateResponse struct {
	ServerTime      int64  `json:"serverTime"`
	Salt            string `json:"salt"`
	AccessToken     string `json:"accessToken"`
	TokenType       string `json:"tokenType"`
	RefreshToken    string `json:"refreshToken"`
	Salt1           int32  `json:"salt1"`
	Salt2           int32  `json:"salt2"`
	Salt3           int32  `json:"salt3"`
	Salt4           int32  `json:"salt4"`
	Salt5           int32  `json:"salt5"`
	IsDisplayActive bool   `json:"isDisplayActive"`
	PopupDocFor     string `json:"popupDocFor"`
}

func (a *AuthenticateResponse) GetParsedAccessToken() string {
	i1 := cdx(a.Salt1, a.Salt2)
	i2 := rdx(a.Salt1, a.Salt2, a.Salt4)
	i3 := bdx(a.Salt1, a.Salt2, a.Salt4)
	i4 := ndx(a.Salt1, a.Salt2, a.Salt4)
	i5 := mdx(a.Salt1, a.Salt2, a.Salt4)
	return a.AccessToken[0:i1] + a.AccessToken[i1+1:i2] + a.AccessToken[i2+1:i3] + a.AccessToken[i3+1:i4] + a.AccessToken[i4+1:i5] + a.AccessToken[i5+1:]
}

func (a *AuthenticateResponse) GetParsedRefreshToken() string {
	i1 := cdx(a.Salt2, a.Salt1)
	i2 := rdx(a.Salt2, a.Salt1, a.Salt3)
	i3 := bdx(a.Salt2, a.Salt1, a.Salt4)
	i4 := ndx(a.Salt2, a.Salt1, a.Salt4)
	i5 := mdx(a.Salt2, a.Salt1, a.Salt4)
	return a.RefreshToken[0:i1] + a.RefreshToken[i1+1:i2] + a.RefreshToken[i2+1:i3] + a.RefreshToken[i3+1:i4] + a.RefreshToken[i4+1:i5] + a.RefreshToken[i5+1:]
}
