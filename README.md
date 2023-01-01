# MFCheck

Use JSON like this.

conf.json:

{
	"connect": {
		"url": "https://[url]/",
		"token": "[token]",
		"pin": "[PIN]",
		"batch": "filename.bat",
		"path": "C:\\PATH\\"
	},
	"check": {
			"process": {
				"param1": "app_name_1",
				"param2": "app_name_2"
			},			
			"device":{
				"camera":{
					"ip":"192.168.14.73",
					"port":"80"
				}
			}
	}
}
