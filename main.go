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
type Province []struct {
	Code string `json:"key"`
	Name string `json:"value"`
}

var valid_contract_value = "ContractNotice"

var myClient = &http.Client{Timeout: 10 * time.Second}
var auction_summary = ArrayOfResponseStructure{}

// tylko ogłoszenia o zamówieniu, reszta out
// order by localisation
// wysłac z linkiem

func main() {
	//sendEmail()
	file_all, err := os.Open("przetargi.txt")
	file_filtered, err := os.Create("filtered_auction.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer func(file_all *os.File) {
		err := file_all.Close()
		if err != nil {
		}
	}(file_all)
	defer func(file_filtered *os.File) {
		err := file_filtered.Close()
		if err != nil {
		}
	}(file_filtered)
	filtered_word := "gawła"
	getAuctionFromGovermentSite(file_all)
	writeFilteredData(file_all, filtered_word, file_filtered)
}

func getAuctionFromGovermentSite(file_all *os.File) {
	counter := 1
	now := time.Now()
	week_ago := now.Add(time.Duration(-10) * time.Hour)
	formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		week_ago.Year(), week_ago.Month(), week_ago.Day(),
		week_ago.Hour(), week_ago.Minute(), week_ago.Second())
	for {
		jsonProvince := &Province{}
		provinceUrl := "https://ezamowienia.gov.pl/mo-board/api/v1/glossary?glossaryType=province"
		err := getJson(provinceUrl, jsonProvince)
		if err != nil {
			break
		}

		jsonResponseAuction := &ArrayOfResponseStructure{}
		auctionUrl := "https://ezamowienia.gov.pl/mo-board/api/v1/Board/Search?publicationDateFrom=" + formatted + "Z&SortingColumnName=PublicationDate&SortingDirection=DESC&PageNumber=" + strconv.Itoa(counter) + "&PageSize=10"
		err = getJson(auctionUrl, jsonResponseAuction)
		if err != nil {
			break
		}
		if len(*jsonResponseAuction) == 0 {
			break
		}
		for i := 0; i < 10; i++ {
			removeInvalidOrderType(jsonResponseAuction)
		}
		addProvinceName(jsonResponseAuction, jsonProvince)
		counter++
		fmt.Println(len(auction_summary))
	}
	//saveToFile(jsonResponse, file_all)
}

func addProvinceName(jsonResponseAuction *ArrayOfResponseStructure, jsonProvince *Province) {
	for _, auction := range *jsonResponseAuction {
		provinceCode := auction.OrganizationProvince
		for _, value := range *jsonProvince {
			if provinceCode == value.Code {
				auction.OrganizationProvince = value.Name
				appendToAuctionSummary(auction)
				break
			}
		}
	}
}

func appendToAuctionSummary(jsonResponse struct {
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
}) {
	auction_summary = append(auction_summary, jsonResponse)
}

func removeInvalidOrderType(jsonResponse *ArrayOfResponseStructure) {
	for index, value := range *jsonResponse {
		if value.NoticeType != valid_contract_value {
			*jsonResponse = RemoveIndex(*jsonResponse, index)
			break
		}
	}
}
func RemoveIndex(s ArrayOfResponseStructure, index int) ArrayOfResponseStructure {
	return append(s[:index], s[index+1:]...)
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
	from := "john.doe@example.com"

	user := "91399c960760af"
	password := "fb7986dd7608c4"

	to := []string{
		"roger.roe@example.com",
	}

	addr := "smtp.mailtrap.io:2525"
	host := "smtp.mailtrap.io"

	msg := []byte("From: john.doe@example.com\r\n" +
		"To: roger.roe@example.com\r\n" +
		"Subject: Test mail\r\n\r\n" +
		"Email body\r\n")

	auth := smtp.PlainAuth("", user, password, host)

	err := smtp.SendMail(addr, auth, from, to, msg)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Email sent successfully")
}
