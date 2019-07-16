package main

// ----- Attachment ----- //
type Attachment struct {
	DocType       string `json:"docType"`
	Id            string `json:"id"`
	FileName      string `json:"fileName"`
	FileHash      string `json:"fileHash"`     // 文件Hash值
	FileType      string `json:"fileType"`     // 文件类型
	Creator       string `json:"creator"`      // 创建人
	LastModifiers string `json:"lastModifier"` // 最后修改人
}
