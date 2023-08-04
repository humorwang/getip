package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/humorwang/getip/src/realip"
	"github.com/oschwald/geoip2-golang"
)

var port string



type ipInfo struct {
	ClientIp string  `json:"clientIp,omitempty"`
	RealIp string `json:"realIp,omitempty"`
  CountryCode string   `json:"countryCode,omitempty"`
	CountryName string    `json:"countryName,omitempty"`
  CityName    string    `json:"cityName,omitempty"`
	Address 		string    `json:"address,omitempty"`
	TimeZone    string    `json:"timeZone,omitempty"`
	Latitude    float64   `json:"latitude,omitempty"`
	Longitude		float64   `json:"longitude,omitempty"`
	Asn         uint    `json:"asn,omitempty"`
	AsName      string    `json:"asName,omitempty"`
}

func buildIpInfo(countryCode, countryName, cityName, address,timeZone string, latitude, longitude float64) ipInfo {
  return ipInfo{
    CountryCode:  countryCode,
    CountryName: countryName,
		CityName: cityName,
		Address: address,
		TimeZone: timeZone,
		Latitude: latitude,
		Longitude: longitude,
  }
}


func getIpInfo(ipStr string, language string)(ipInfo , error) {
	db, err := geoip2.Open("./geolite2/GeoLite2-City.mmdb")

	ipInfo := ipInfo{}
	if err != nil {
		log.Fatal(err)
		if language == "zh" {
			return ipInfo, errors.New("ip 数据库加载异常")
		} else {
			return ipInfo, errors.New("load ip db error")
		}
	}
	defer db.Close()
	
	// If you are using strings that may be invalid, check that ip is not nil
	ip := net.ParseIP(ipStr)
	record, err := db.City(ip)
	if err != nil {
		log.Fatal(err)
		if language == "zh" {
			return ipInfo, errors.New("获取ip信息失败")
		} else {
			return ipInfo, errors.New("get ip info error")
		}
	}
	countryCode :=  record.Country.IsoCode
	cityName := record.City.Names["en"]
	countryName := record.Country.Names["en"]
	if countryCode == "CN" {
		cityName = record.City.Names["zh-CN"]
		if len(record.Subdivisions) > 0 {
			cityName = fmt.Sprintf("%v %v", record.Subdivisions[0].Names["zh-CN"], cityName)
		}
		countryName = record.Country.Names["zh-CN"]
	} else {
		if len(record.Subdivisions) > 0 {
			cityName = fmt.Sprintf("%v %v", record.Subdivisions[0].Names["en"], cityName)
		}
	}
	// fmt.Printf("Portuguese (BR) city name: %v\n", record.City.Names["zh-CN"])
	// if len(record.Subdivisions) > 0 {
	// 	fmt.Printf("English subdivision name: %v\n", record.Subdivisions[0].Names["en"])
	// }
	// fmt.Printf("Russian country name: %v\n", record.Country.Names["en"])
	// fmt.Printf("ISO country code: %v\n", record.Country.IsoCode)
	// fmt.Printf("Time zone: %v\n", record.Location.TimeZone)
	timeZone := record.Location.TimeZone
	latitude := record.Location.Latitude
	longitude := record.Location.Longitude
	fmt.Printf("Coordinates: %v, %v\n", record.Location.Latitude, record.Location.Longitude)
	address := fmt.Sprintf("%v %v %v", cityName, countryName, countryCode)
	asNumber , asName := track_asn(ipStr)

	ipInfo = buildIpInfo(countryCode, countryName, cityName , address , timeZone, latitude, longitude)
	ipInfo.Asn = asNumber
	ipInfo.AsName = asName

	return ipInfo, nil
}


func track_asn(ipStr string)(uint, string){
	db, err := geoip2.Open("./geolite2/GeoLite2-ASN.mmdb")
	if err != nil {
		log.Fatal(err)
		return 0, ""
	}
	defer db.Close()
	ip := net.ParseIP(ipStr)
	record, err :=db.ASN(ip)
	if err != nil {
		log.Fatal(err)
		return 0, ""
	}
	fmt.Printf("ASN number: %v\n", record.AutonomousSystemNumber)
	fmt.Printf("ASN name: %v\n", record.AutonomousSystemOrganization)
	return record.AutonomousSystemNumber, record.AutonomousSystemOrganization
}







func home(c *gin.Context) {
		// log.Printf("\n============================================================================\n"+
	// 	"IP:%s\n"+
	// 	"X-Forwarded-For:%s\n"+
	// 	"X-Real-Ip:%s\n"+
	// 	"X-Forwarded-Host:%s\n"+
	// 	"RemoteAddr:%s\n"+
	// 	"Content-Type:%s\n"+
	// 	"IpInfo:%s\n",
	// 	c.ClientIP(),
	// 	c.Request.Header.Get("X-Forwarded-For"),
	// 	c.Request.Header.Get("X-Real-Ip"),
	// 	c.Request.Header.Get("X-Forwarded-Host:"),
	// 	c.Request.RemoteAddr,
	// 	c.Request.Header.Get("Content-Type"),
	// 	getIpInfo(ip))
	format := c.DefaultQuery("format", "text")
	language := c.DefaultQuery("language", "zh")
	ip := c.ClientIP()
	realIp := realip.FromRequest(c.Request)
	errorInfo := ""
	isPrivate, err := realip.IsPrivateAddress(realIp)
	if err != nil {
		if language == "zh" {
			errorInfo = "检查局域网错误：" + err.Error()
		} else {
			errorInfo = "check private error: " + err.Error()
		}
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{"clientIp": ip, "realIp": realIp, "address": errorInfo})
		} else {
			c.String(http.StatusOK, fmt.Sprintf("%v %v %v", ip, realIp, errorInfo))
		}
		return
	}
	if isPrivate {
		if language == "zh" {
			errorInfo = "局域网地址"
		} else {
			errorInfo = "LAN address"
		}
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{"clientIp": ip, "realIp": realIp, "address": errorInfo})
		} else {
			c.String(http.StatusOK, fmt.Sprintf("%v %v %v", ip, realIp, errorInfo))
		}
		return
	}

	ipInfo, err := getIpInfo(realIp, language)
	if err != nil {
		if format == "json" {
			c.JSON(http.StatusOK, gin.H{"clientIp": ip, "realIp": realIp, "address": err.Error()})
		} else {
			c.String(http.StatusOK, fmt.Sprintf("%v %v %v", ip, realIp, err.Error()))
		}
		return
	}
	ipInfo.ClientIp = ip
	ipInfo.RealIp = realIp
	if format == "json" {
		c.JSON(http.StatusOK, ipInfo)
	} else {
		c.String(http.StatusOK, fmt.Sprintf("%v %v", realIp, ipInfo.Address))
	}
}

func get_ip_info(c *gin.Context) {
	ip := c.Param("ip")
	language := c.DefaultQuery("language", "zh")
	ipInfo, err := getIpInfo(ip, language)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"ip": ip, "address": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ipInfo)
}



func init() {
	flag.StringVar(&port, "port", ":8080", "端口")
}



func main() {
	flag.Parse()
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/", home)
	r.GET("/ip/:ip", get_ip_info)

	r.Run(port)
}
