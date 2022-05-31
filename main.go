package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type ArrayOfResponseStructure []struct {
	ClientType                  string      `json:"clientType"`
	OrderType                   interface{} `json:"orderType"`
	TenderType                  string      `json:"tenderType"`
	NoticeType                  string      `json:"noticeType"`
	NoticeTypeDisplayName       interface{} `json:"noticeTypeDisplayName"`
	NoticeNumber                string      `json:"noticeNumber"`
	BzpNumber                   string      `json:"bzpNumber"`
	IsTenderAmountBelowEU       bool        `json:"isTenderAmountBelowEU"`
	PublicationDate             time.Time   `json:"publicationDate"`
	OrderObject                 string      `json:"orderObject"`
	CpvCode                     string      `json:"cpvCode"`
	SubmittingOffersDate        interface{} `json:"submittingOffersDate"`
	ProcedureResult             interface{} `json:"procedureResult"`
	OrganizationName            string      `json:"organizationName"`
	OrganizationCity            string      `json:"organizationCity"`
	OrganizationProvince        string      `json:"organizationProvince"`
	OrganizationCountry         string      `json:"organizationCountry"`
	OrganizationNationalID      string      `json:"organizationNationalId"`
	UserID                      string      `json:"userId"`
	OrganizationID              string      `json:"organizationId"`
	MoIdentifier                string      `json:"moIdentifier"`
	TenderID                    string      `json:"tenderId"`
	IsManuallyLinkedWithTender  bool        `json:"isManuallyLinkedWithTender"`
	HTMLBody                    interface{} `json:"htmlBody"`
	Contractors                 interface{} `json:"contractors"`
	BzpTenderPlanNumber         interface{} `json:"bzpTenderPlanNumber"`
	BaseNoticeMOIdentifier      string      `json:"baseNoticeMOIdentifier"`
	TechnicalNoticeMOIdentifier interface{} `json:"technicalNoticeMOIdentifier"`
	Outdated                    bool        `json:"outdated"`
	ObjectID                    string      `json:"objectId"`
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func main() {
	sendEmailByGoogle()
	//file_all, err := os.Open("przetargi.txt")
	//file_filtered, err := os.Create("filtered_auction.txt")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer file_all.Close()
	//defer file_filtered.Close()
	//filtered_word := "gawła"
	//getAuctionFromGovermentSite(file_all)
	//writeFilteredData(file_all, filtered_word, file_filtered)
}

func getAuctionFromGovermentSite(file_all *os.File) {
	counter := 1
	for {
		jsonResponse := &ArrayOfResponseStructure{}
		url := "https://ezamowienia.gov.pl/mo-board/api/v1/Board/Search?publicationDateFrom=2022-04-30T13:07:46.343Z&SortingColumnName=PublicationDate&SortingDirection=DESC&PageNumber=" + strconv.Itoa(counter) + "&PageSize=10"
		err := getJson(url, jsonResponse)
		if err != nil {
			break
		}
		if len(*jsonResponse) == 0 {
			break
		}
		saveToFile(jsonResponse, file_all)
		counter++
	}
}

func writeFilteredData(file_all *os.File, filtered_word string, file_filtered *os.File) {
	scanner := bufio.NewScanner(file_all)
	line := 1
	for scanner.Scan() {
		if strings.Contains(strings.ToUpper(scanner.Text()), strings.ToUpper(filtered_word)) {
			b, err := fmt.Fprintln(file_filtered, scanner.Text()+"\n")
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%d bytes written successfully to file: "+file_filtered.Name()+"!\n", b)
		}

		line++
	}

	if err := scanner.Err(); err != nil {
		// Handle the error
	}
}

func saveToFile(jsonResponse *ArrayOfResponseStructure, file *os.File) {
	for _, value := range *jsonResponse {

		b, err := fmt.Fprintln(file, value.OrderObject+"\n")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%d bytes written successfully to file: "+file.Name()+"\n", b)
	}
}
func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func sendEmailByGoogle() {
	from := "91399c960760af"
	password := "fb7986dd7608c4"

	toEmailAddress := "m.kapela91@gmail.com"
	to := []string{toEmailAddress}

	host := "smtp.mailtrap.io"
	port := "2525"
	address := host + ":" + port

	subject := "Subject: This is the subject of the mail\n"
	body := "This is the body of the mail"
	message := []byte(subject + body)

	auth := smtp.PlainAuth("", from, password, host)

	err := smtp.SendMail(address, auth, from, to, message)
	if err != nil {
		panic(err)
	}
}
