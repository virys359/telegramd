package auth

import (
	"time"
	"github.com/nebulaim/telegramd/biz/dal/dataobject"
	"github.com/nebulaim/telegramd/biz/dal/dao"
	"fmt"
	"github.com/nebulaim/telegramd/biz/base"
	"github.com/golang/glog"
	"github.com/nebulaim/telegramd/mtproto"
)

type phoneCodeData struct {
	phoneNumber string
	Code string
	length int32
	CodeHash   string
}

func (code* phoneCodeData) GetPhoneCodeLength() int32 {
	// TODO(@benqi): 6???
	return 6
}

func GetOrCreateCode(authKeyId int64, phoneNumber string, apiId int32, apiHash string) *phoneCodeData {
	// 15分钟内有效
	// var code *codeData
	_ = authKeyId

	lastCreatedAt := time.Unix(time.Now().Unix()-15*60, 0).Format("2006-01-02 15:04:05")
	do := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE).SelectByPhoneAndApiIdAndHash(phoneNumber, apiId, apiHash, lastCreatedAt)
	if do == nil {
		do = &dataobject.AuthPhoneTransactionsDO{
			ApiId: apiId,
			ApiHash: apiHash,
			PhoneNumber: phoneNumber,
			// TODO(@benqi): gen rand number
			Code: "123456",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			// TODO(@benqi): 生成一个32字节的随机字串
			TransactionHash:fmt.Sprintf("%20d", base.NextSnowflakeId()),
		}
		dao.GetAuthPhoneTransactionsDAO(dao.DB_MASTER).Insert(do)
	} else {
		// TODO(@benqi): FLOOD_WAIT_X, too many attempts, please try later.
	}

	code := &phoneCodeData{
		phoneNumber: phoneNumber,
		Code:        do.Code,
		length:      6,
		CodeHash:    do.TransactionHash,
	}

	return code
}

func GetCodeByCodeHash(phoneNumber, codeHash string) (*phoneCodeData, error) {
	do := dao.GetAuthPhoneTransactionsDAO(dao.DB_SLAVE).SelectByPhoneCodeHash(codeHash, phoneNumber)
	if do == nil {
		err := mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_HASH_EMPTY), "invalid phone number")
		glog.Error(err)
		return nil, err
	}

	if time.Now().Unix() > do.CreatedTime + 15*60 {
		err := mtproto.NewRpcError(int32(mtproto.TLRpcErrorCodes_PHONE_CODE_EXPIRED), "code expired")
		glog.Error(err)
		return nil, err
	}

	if do.Attempts > 3 {
		// TODO(@benqi): 输入了太多次错误的phone code
		err := mtproto.NewFloodWaitX(15*60, "auth.sendCode#86aef0ec: too many attempts.")
		return nil, err
	}

	code := &phoneCodeData{
		phoneNumber: phoneNumber,
		Code:        do.Code,
		length:      6,
		CodeHash:    codeHash,
	}
	return code, nil
}
