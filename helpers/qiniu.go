package helpers

import (
	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
)

const STATICBUCKET = "static"
const AUDIOBUCKET = "course-audio"

var QiniuAK string
var QiniuSK string

type SmQiniu struct {
	Ak string
	Sk string
}

func GetQiniu() *SmQiniu {
	return &SmQiniu{
		QiniuAK,
		QiniuSK,
	}
}

// 获取静态文件的上传凭证
func (qn *SmQiniu) GetStaticFileUpToken() string {
	putPolicy := storage.PutPolicy{
		Scope: STATICBUCKET,
	}
	mac := qbox.NewMac(qn.Ak, qn.Sk)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}

// 获取音频文件的上传凭证
func (qn *SmQiniu) GetAudioFileUpToken() string {
	putPolicy := storage.PutPolicy{
		Scope:      AUDIOBUCKET,
		ReturnBody: `{"key":"$(key)","hash":"$(etag)","fsize":$(fsize),"fname":$(fname),"duration":$(avinfo.audio.duration)}`,
	}
	mac := qbox.NewMac(qn.Ak, qn.Sk)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}

// ueditor上传凭证
func (qn *SmQiniu) GetUEditorImageUpToken() string {
	putPolicy := storage.PutPolicy{
		Scope:      STATICBUCKET,
		ReturnBody: `{"state":"SUCCESS","key":"$(key)","hash":"$(etag)","url":"$(key)"}`,
	}
	mac := qbox.NewMac(qn.Ak, qn.Sk)
	upToken := putPolicy.UploadToken(mac)
	return upToken
}
