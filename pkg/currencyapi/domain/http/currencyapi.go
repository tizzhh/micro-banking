package currencyapihttp

import "encoding/json"

type Reponse struct {
	Meta Meta `json:"meta"`
	Data Data `json:"data"`
}

type Meta struct {
	LastUpdated string `json:"last_updated_at"`
}

type Data struct {
	Currencies map[string]Currency `json:"-"`
}

type Currency struct {
	Code  string  `json:"code"`
	Value float32 `json:"value"`
}

func (d *Data) UnmarshalJSON(data []byte) error {
	temp := make(map[string]Currency)

	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	d.Currencies = temp
	return nil
}
