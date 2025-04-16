package tokens

func (tf *TokenFactory) NewTokenPayload() *TokenPayload {
	return &TokenPayload{
		builder: nil,
		tf:      tf,
	}
}

func (p *TokenPayload) SetID(id string) *TokenPayload {
	p.ID = id
	return p
}

func (p *TokenPayload) SetUsername(username string) *TokenPayload {
	p.Username = username
	return p
}

func (p *TokenPayload) SetEmail(email string) *TokenPayload {
	p.Email = email
	return p
}

func (p *TokenPayload) SetExtraClaims(extraClaims map[string]interface{}) *TokenPayload {
	p.ExtraClaims = extraClaims
	return p
}

func (p *TokenPayload) SetExtraClaim(key string, value interface{}) *TokenPayload {
	p.ExtraClaims[key] = value
	return p
}

func (p *TokenPayload) ToMap() map[string]interface{} {
	return map[string]interface{}{
		"id":           p.ID,
		"username":     p.Username,
		"email":        p.Email,
		"extra_claims": p.ExtraClaims,
	}
}
