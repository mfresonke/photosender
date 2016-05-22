package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	twilio "github.com/carlosdp/twiliogo"
)

const senderNum = "4074998165"

const (
	baseDir = "/Users/max/Desktop/TaylorBDayProject"
	csvLoc  = baseDir + "/Everyone.csv"
)

const (
	accountSid       = "acctSidGoesHere"
	accountAuthToken = "authTokenGoesHere"
)

const (
	imgURLPrefix  = "http://e11e606e.ngrok.io/IMG_"
	imgURLPostfix = ".JPG"
)

var numToPeep = make(map[int]peep)

func main() {
	// Load CSV into Memory
	parsePeeps()
	recips, photoURLs := parseArgs()
	log.Println("Sending", len(photoURLs), "pictures to", len(recips), "recipients!")
	log.Println("This request will take approximately", len(recips)*len(photoURLs)+len(recips), "seconds!")
	log.Println("Debug: ", recips, photoURLs)
	client := twilio.NewClient(accountSid, accountAuthToken)

	var msgs []*twilio.Message
	//send warning sms
	for _, recip := range recips {
		if len(recip) < 10 {
			continue
		}
		message, err := twilio.NewMessage(client, senderNum, recip, twilio.Body("Pictures Sent for Taylor's 15th Birthday. Reply STOP to opt-out."))
		if err != nil {
			log.Println("Error sending to: ", recip)
			log.Println(err)
		}
		msgs = append(msgs, message)
		time.Sleep(time.Second)
	}

	for _, photoURL := range photoURLs {
		for _, recip := range recips {
			if len(recip) < 10 {
				continue
			}
			message, err := twilio.NewMessage(client, senderNum, recip, twilio.MediaUrl(photoURL))
			if err != nil {
				log.Println("Error sending to: ", recip)
				log.Println(err)
			}
			msgs = append(msgs, message)
			time.Sleep(time.Second)
		}
	}
	// debug print msgs
	// for _, msg := range msgs {
	// 	log.Println(msg)
	// }
}

type peep struct {
	name        string
	phoneNumber string
	id          int
}

func parsePeeps() {
	file, err := os.Open(csvLoc)
	check(err)
	csvReader := csv.NewReader(file)
	records, err := csvReader.ReadAll()
	check(err)
	//cool, now we have records. Let's make a map!
	for recordNum, record := range records {
		if recordNum == 0 {
			continue
		}
		idStr := record[0]
		id, err := strconv.Atoi(idStr)
		check(err)
		name := record[1]
		phoneNum := record[2]
		numToPeep[id] = peep{
			name:        name,
			phoneNumber: phoneNum,
			id:          id,
		}
	}
}

// parse args needs to be able to handle single comma-separated vals,
// but also dash separated vals within that, e.g. 12423-12426
func parseArgs() (recipients, photoURLs []string) {
	if len(os.Args) != 3 {
		panic("Proper Usage: photosender <Nums to Send to> <Nums Of Photos>")
	}
	argRecips := os.Args[1]
	argPics := os.Args[2]

	// if everyone, then send to everyone in the map
	if argRecips == "all" {
		for _, peep := range numToPeep {
			recipients = append(recipients, peep.phoneNumber)
		}
	} else {
		recipNums := parseNumList(argRecips)
		// now to turn the slice of recipient numbers into phone numbers we can use
		for _, recipNum := range recipNums {
			peep, ok := numToPeep[recipNum]
			if !ok {
				panic("Recipient not found!")
			}
			recipients = append(recipients, peep.phoneNumber)
			log.Println("Adding ", peep.name, " to the list of recipients")
		}
	}

	//now let's do something similar for the photos
	photoNums := parseNumList(argPics)
	for _, photoNum := range photoNums {
		photoURL := imgURL(photoNum)
		photoURLs = append(photoURLs, photoURL)
	}
	return recipients, photoURLs
}

func parseNumList(numList string) (nums []int) {
	commaSplitNums := strings.Split(numList, ",")
	for _, rawRange := range commaSplitNums {
		// if the raw recip is a range
		if strings.Contains(rawRange, "-") {
			// separate out the range into individual numbers
			numRange := fillRange(rawRange)
			nums = append(nums, numRange...)
		} else {
			num, err := strconv.Atoi(rawRange)
			check(err)
			nums = append(nums, num)
		}
	}
	return nums
}

func fillRange(inRange string) (outRange []int) {
	if !strings.Contains(inRange, "-") {
		panic("invalid number range!")
	}
	// separate out the range into individual numbers
	recipRange := strings.Split(inRange, "-")
	beg, err := strconv.Atoi(recipRange[0])
	check(err)
	end, err := strconv.Atoi(recipRange[1])
	check(err)
	if end < beg {
		panic("invalid number range!")
	}
	for i := beg; i <= end; i++ {
		outRange = append(outRange, i)
	}
	return outRange
}

func imgURL(imgNum int) string {
	//basic sanity check
	if imgNum < 0 || imgNum > 100000 {
		panic("imgNum invalid")
	}
	return fmt.Sprint(imgURLPrefix, imgNum, imgURLPostfix)
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
