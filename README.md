## Install

```bash
cd parser_letu
export GOPATH=`pwd`
go get golang.org/x/net/html
go get gopkg.in/mgo.v2
go get github.com/blackjack/syslog
# goldapple pkg
go build goldapple
go install goldapple
go run tcp-server.go
```

## Cron

```bash
crontab -e
```

```bash
# everyday at 11:00

00 11 * * * echo -n "start"|netcat 127.0.0.1 8800
```

### Docker

```bash
# Start
cd parser-letu
sudo docker build -t ga/parser-letu .
# !!! network host -> localhost MongoDB
sudo docker run --network host -d --restart always --log-driver syslog gapple/parser-letu:latest
# Stop
sudo docker ps
sudo docker kill <image_name>
```

```bash

db['ILDE_products_final'].aggregate([
    {
        $match: {
        	gestori: {
            	$gt: ""
        	}
        }
    },
	{
		$lookup: {
			from: "gestori_up",
			localField: "gestori",
			foreignField: "Cod_good",
			as: "sku"
		}
	},
	{
		$out: "gestori_export"
	}
])

db['ILDE_products_final'].aggregate([
    {
        $match: {
        	gestori: {
            	$gt: ""
        	}
        }
    },
    {
        $match: {
        	gestori: {
            	$ne: ""
        	}
        }
    },
	{
		$lookup: {
			from: "gestori_rc",
			localField: "gestori",
			foreignField: "Cod_good",
			as: "sku"
		}
	},
	{
		$project: {
			url: 1,
			Navi: 1,
			Brand: 1,
			articul: 1,
			Barcod: '$sku.Barcod',
			Name: '$sku.Name'
		}
	},
	{
		$out: "gest_rc_1"
	}
])

db['ILDE_products_final'].aggregate([
    {
        $match: {
        	gestori: {
            	$gt: ""
        	}
        }
    },
    {
        $match: {
        	gestori: {
            	$ne: ""
        	}
        }
    },
	{
		$lookup: {
			from: "gestori_up",
			localField: "gestori",
			foreignField: "Cod_good",
			as: "sku"
		}
	},
	{
		$project: {
			url: 1,
			Navi: 1,
			Brand: 1,
			articul: 1,
			Barcod: '$sku.Barcod',
			Name: '$sku.Name'
		}
	},
	{
		$out: "gest_exp"
	}
])

mongoexport --host localhost --username apidev --password "apidev" --collection gestori_export --db parser --out /home/administrator/exp.csv --type csv --fields url, articul, sku.Barcod


```
