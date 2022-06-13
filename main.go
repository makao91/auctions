package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
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
var ssp_auctions = ArrayOfResponseStructure{}
var keywords_for_request = []string{
	"sygnalizacji%20po%C5%BCar",
	"ssp",
	"cctv",
}
var keywords_for_sorting = map[string][]string{
	"ssp":  {"ssp"},
	"cctv": {"cctv"},
	"sygnalizacji%20po%C5%BCar": {
		"sygnalizacji",
		"pożar",
	},
}

// tylko ogłoszenia o zamówieniu, reszta out
// order by localisation
// wysłac z linkiem
// przykład filtra po słowach: SSP, CCTV, sygnalizacji pożaru

func main() {
	for _, keyword := range keywords_for_request {
		file, err := os.Create(keyword + "_przetargi.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer func(file_all *os.File) {
			err := file_all.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(file)

		getAuctionFromGovermentSite(file, keyword)
	}
	//sendEmail()
}

func getAuctionFromGovermentSite(file_all *os.File, keyword string) {
	break_download_new_auctions := false
	counter := 1
	now := time.Now()
	week_ago := now.Add(time.Duration(-720) * time.Hour)
	formatted := fmt.Sprintf("%d-%02d-%02dT%02d:%02d:%02d",
		week_ago.Year(), week_ago.Month(), week_ago.Day(),
		week_ago.Hour(), week_ago.Minute(), week_ago.Second())
	jsonProvince := &Province{}

	for {
		err := loadProvinces(jsonProvince)
		if err != nil {
			log.Fatal(err)
		}

		jsonResponseAuction := &ArrayOfResponseStructure{}
		err = loadAuctions(keyword, formatted, counter, err, jsonResponseAuction)

		if len(*jsonResponseAuction) == 0 {
			break
		}
		for i := 0; i < 10; i++ {
			removeInvalidOrderType(jsonResponseAuction)
		}
		addProvinceName(jsonResponseAuction, jsonProvince)
		counter++
		break_download_new_auctions = isPublicationDateInTheRangeOfTwoWeeks(jsonResponseAuction, week_ago, break_download_new_auctions)
		if break_download_new_auctions == true {
			break
		}
		//fmt.Println(len(auction_summary))
	}
	keyword_counter := len(keywords_for_sorting[keyword])
	for _, auction := range auction_summary {
		switch keyword_counter {
		case 1:
			if strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][0]) == true {
				appendToSspSummary(auction)
			}
		case 2:
			if strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][0]) == true && strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][1]) == true {
				appendToSspSummary(auction)
			}
		case 3:
			if strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][0]) == true && strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][1]) == true && strings.Contains(strings.ToLower(auction.OrderObject), keywords_for_sorting[keyword][2]) == true {
				appendToSspSummary(auction)
			}
		}

	}
	sort.Slice(ssp_auctions, func(i, j int) bool {
		return ssp_auctions[i].OrganizationProvince < ssp_auctions[j].OrganizationProvince
	})
	//strings.Contains("something", "some")
	saveToFile(&ssp_auctions, file_all)
}

func isPublicationDateInTheRangeOfTwoWeeks(jsonResponseAuction *ArrayOfResponseStructure, week_ago time.Time, break_download_new_auctions bool) bool {
	for _, auction := range *jsonResponseAuction {
		if week_ago.After(auction.PublicationDate) {
			break_download_new_auctions = true
			break
		}
	}
	return break_download_new_auctions
}

func loadAuctions(keyword string, formatted string, counter int, err error, jsonResponseAuction *ArrayOfResponseStructure) error {
	auctionUrl := "https://ezamowienia.gov.pl/mo-board/api/v1/Board/Search?keyword=" + keyword + "&publicationDateFrom=" + formatted + "Z&SortingColumnName=PublicationDate&SortingDirection=DESC&PageNumber=" + strconv.Itoa(counter) + "&PageSize=10"
	err = getJson(auctionUrl, jsonResponseAuction)
	return err
}

func loadProvinces(jsonProvince *Province) error {
	provinceUrl := "https://ezamowienia.gov.pl/mo-board/api/v1/glossary?glossaryType=province"
	err := getJson(provinceUrl, jsonProvince)
	return err
}

func appendToSspSummary(auction struct {
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
	ssp_auctions = append(ssp_auctions, auction)
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

//func writeFilteredData(file_all *os.File, filtered_word string, file_filtered *os.File) {
//	scanner := bufio.NewScanner(file_all)
//	line := 1
//	for scanner.Scan() {
//		if strings.Contains(strings.ToUpper(scanner.Text()), strings.ToUpper(filtered_word)) {
//			b, err := fmt.Fprintln(file_filtered, scanner.Text()+"\n")
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			fmt.Printf("%d bytes written successfully to file: "+file_filtered.Name()+"!\n", b)
//		}
//
//		line++
//	}
//
//	if err := scanner.Err(); err != nil {
//		log.Fatal(err)
//	}
//}

func saveToFile(jsonResponse *ArrayOfResponseStructure, file *os.File) {
	for _, value := range *jsonResponse {

		b, err := fmt.Fprintln(file,
			value.OrderObject+"\n",
			"Województwo: "+value.OrganizationProvince+"\n",
			"Miasto: "+value.OrganizationCity+"\n",
			"Link: "+"https://ezamowienia.gov.pl/mo-client-board/bzp/notice-details/id/"+value.ObjectID+"\n",
			"Data publikacji: "+value.PublicationDate.String()+"\n",
		)
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
