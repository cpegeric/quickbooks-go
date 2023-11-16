package quickbooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
)

type Purchase struct {
	ID                      string `json:"Id,omitempty"`
	Line                    []Line
	PaymentType             string           `json:",omitempty"`
	AccountRef              *ReferenceType   `json:",omitempty"`
	SyncToken               string           `json:",omitempty"`
	CurrencyRef             *ReferenceType   `json:",omitempty"`
	TxnDate                 *Date            `json:",omitempty"`
	PrintStatus             string           `json:",omitempty"` // Valid values: NotSet, NeedToPrint, PrintComplete
	RemitToAddr             *PhysicalAddress `json:",omitempty"`
	TxnStatus               string           `json:",omitempty"`
	GlobalTaxCalculation    string           `json:",omitempty"` // Values: TaxExcluded, TaxInclusive, NotApplicable
	TransactionLocationType string           `json:",omitempty"`
	MetaData                *MetaData        `json:",omitempty"`
	DocNumber               string           `json:",omitempty"`
	PrivateNote             string           `json:",omitempty"`
	Credit                  bool             `json:",omitempty"`
	TxnTaxDetail            *TxnTaxDetail    `json:",omitempty"`
	PaymentMethodRef        *ReferenceType   `json:",omitempty"`
	ExchangeRate            json.Number      `json:",omitempty"`
	DepartmentRef           *ReferenceType   `json:",omitempty"`
	EntityRef               *ReferenceType   `json:",omitempty"`
	IncludeInAnnualTPAR     bool             `json:",omitempty"`
	TotalAmt                json.Number      `json:",omitempty"`
	RecurDataRef            *ReferenceType   `json:",omitempty"`
}

func (c *Client) CreatePurchase(purchase *Purchase) (*Purchase, error) {
	var u, err = url.Parse(string(c.Endpoint))
	if err != nil {
		return nil, err
	}
	u.Path = "/v3/company/" + c.RealmID + "/purchase"
	var v = url.Values{}
	v.Add("minorversion", minorVersion)
	u.RawQuery = v.Encode()
	var j []byte
	j, err = json.Marshal(purchase)
	if err != nil {
		return nil, err
	}
	var req *http.Request
	req, err = http.NewRequest("POST", u.String(), bytes.NewBuffer(j))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	var res *http.Response
	res, err = c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, parseFailure(res)
	}

	var r struct {
		Purchase Purchase
		Time     Date
	}

	err = json.NewDecoder(res.Body).Decode(&r)
	return &r.Purchase, err
}

// DeleteInvoice deletes the given Invoice by ID and sync token from the
// QuickBooks server.
func (c *Client) DeletePurchase(id, syncToken string) error {
	var u, err = url.Parse(string(c.Endpoint))
	if err != nil {
		return err
	}
	u.Path = "/v3/company/" + c.RealmID + "/purchase"
	var v = url.Values{}
	v.Add("minorversion", minorVersion)
	v.Add("operation", "delete")
	u.RawQuery = v.Encode()
	var j []byte
	j, err = json.Marshal(struct {
		ID        string `json:"Id"`
		SyncToken string
	}{
		ID:        id,
		SyncToken: syncToken,
	})
	if err != nil {
		return err
	}
	var req *http.Request
	req, err = http.NewRequest("POST", u.String(), bytes.NewBuffer(j))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	var res *http.Response
	res, err = c.Client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	//var b, _ = ioutil.ReadAll(res.Body)
	//log.Println(string(b))

	// If the invoice was already deleted, QuickBooks returns 400 :(
	// The response looks like this:
	// {"Fault":{"Error":[{"Message":"Object Not Found","Detail":"Object Not Found : Something you're trying to use has been made inactive. Check the fields with accounts, invoices, items, vendors or employees.","code":"610","element":""}],"type":"ValidationFault"},"time":"2018-03-20T20:15:59.571-07:00"}

	// This is slightly horrifying and not documented in their API. When this
	// happens we just return success; the goal of deleting it has been
	// accomplished, just not by us.
	if res.StatusCode == http.StatusBadRequest {
		var r Failure
		err = json.NewDecoder(res.Body).Decode(&r)
		if err != nil {
			return err
		}
		if r.Fault.Error[0].Message == "Object Not Found" {
			return nil
		}
	}
	if res.StatusCode != http.StatusOK {
		return parseFailure(res)
	}

	// TODO they send something back, but is it useful?
	return nil
}
