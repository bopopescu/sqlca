// Code generated by db2go. DO NOT EDIT.
package dataobject

var TableNameUsers = "users" //

type UsersDO struct {
	Id        int32   `json:"id" db:"id" `                                 //auto inc id
	Name      string  `json:"name" db:"name" `                             //user name
	Phone     string  `json:"phone" db:"phone" `                           //phone number
	Sex       int8    `json:"sex" db:"sex" `                               //user sex
	Email     string  `json:"email" db:"email" `                           //email
	Disable   int8    `json:"disable" db:"disable" `                       //disabled(0=false 1=true)
	Balance   float64 `json:"balance" db:"balance" `                       //balance of decimal
	CreatedAt string  `json:"created_at" db:"created_at" sqlca:"readonly"` //create time
	UpdatedAt string  `json:"updated_at" db:"updated_at" sqlca:"readonly"` //update time
}

func (do *UsersDO) GetId() int32         { return do.Id }
func (do *UsersDO) SetId(v int32)        { do.Id = v }
func (do *UsersDO) GetName() string      { return do.Name }
func (do *UsersDO) SetName(v string)     { do.Name = v }
func (do *UsersDO) GetPhone() string     { return do.Phone }
func (do *UsersDO) SetPhone(v string)    { do.Phone = v }
func (do *UsersDO) GetSex() int8         { return do.Sex }
func (do *UsersDO) SetSex(v int8)        { do.Sex = v }
func (do *UsersDO) GetEmail() string     { return do.Email }
func (do *UsersDO) SetEmail(v string)    { do.Email = v }
func (do *UsersDO) GetDisable() int8     { return do.Disable }
func (do *UsersDO) SetDisable(v int8)    { do.Disable = v }
func (do *UsersDO) GetBalance() float64  { return do.Balance }
func (do *UsersDO) SetBalance(v float64) { do.Balance = v }
func (do *UsersDO) GetCreatedAt() string { return do.CreatedAt }
func (do *UsersDO) GetUpdatedAt() string { return do.UpdatedAt }
