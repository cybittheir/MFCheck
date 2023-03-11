# MFCheck

_Use JSON like this._

## conf.json:
```
{
	"connect": {
		"url": "https://[url]/",
		"token": "[token]",
		"pin": "[PIN]",
		"batch": "filename.bat",
		"path": "C:\\PATH\\",
		"period": "61" [sec]
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
```
## Options:
* -s (OR -silent): hiding all messages except errors
* -h (OR -help): this message

_Ctrl+C for exit_