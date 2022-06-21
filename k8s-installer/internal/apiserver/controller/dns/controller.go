package dns

import (
	"encoding/json"
	"fmt"
	"k8s-installer/internal/apiserver/utils"
	"k8s-installer/pkg/cache"
	"k8s-installer/pkg/constants"
	"k8s-installer/pkg/coredns"
	"k8s-installer/pkg/log"
	"net/http"
	"strings"

	"github.com/emicklei/go-restful"
)

func ListAllSubDomain(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	result, err := runtimeCache.GetAllSubDomainList()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Errorf("Failed to list all sub domain due to error: %s ", err.Error()).Error())
		return
	}
	domains := []coredns.DNSDomain{}
	for _, domain := range result {
		domains = append(domains, domain)
	}
	response.WriteAsJson(domains)
}

func ListSubDomainOfTLd(request *restful.Request, response *restful.Response) {
	domain := request.PathParameter("domain")
	if len(strings.TrimSpace(domain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Errorf("Path parameter domain is not found ").Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	result, err := runtimeCache.GetSubDomainOfTls(domain)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Errorf("Failed to list all sub domain due to error: %s ", err.Error()).Error())
		return
	}

	if result == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Errorf("Unable to find domain with name %s ", domain).Error())
		return
	}

	domains := []coredns.DNSDomain{}
	for _, domain := range result {
		domains = append(domains, domain)
	}
	response.WriteAsJson(domains)
}

func CreateSubDomain(request *restful.Request, response *restful.Response) {
	domain := request.PathParameter("domain")
	if len(strings.TrimSpace(domain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprint("Path parameter domain is not found "))
		return
	}

	domainPost := coredns.DNSDomain{}
	if err := request.ReadEntity(&domainPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	if domain != domainPost.TopLevelDomain {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Path parameter domain value doest not match post domain TopLevelDomain value"))
		return
	}

	if !strings.HasSuffix(domainPost.Domain, domainPost.TopLevelDomain) {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Domain does not end with top level domain"))
		return
	}

	if strings.HasPrefix(domainPost.Domain, "*.") {
		// if start with *. wo consider it's a wild card
		// first remove *.
		// then do regular fqdn validation
		copyDomainPost := domainPost
		copyDomainPost.Domain = copyDomainPost.Domain[2:]
		if err := utils.DnsValidate(copyDomainPost); err != nil {
			utils.ResponseError(response, http.StatusBadRequest, err.Error())
			return
		}
	} else {
		// do sub domain name validation check
		if err := utils.DnsValidate(domainPost); err != nil {
			utils.ResponseError(response, http.StatusBadRequest, err.Error())
			return
		}
	}

	// do resolve validation
	for index, domain := range domainPost.DomainResolve {
		domainPost.DomainResolve[index].ResolveDomain = domainPost.Domain
		err := utils.DnsValidate(domain)
		if err != nil {
			utils.ResponseError(response, http.StatusBadRequest, err.Error())
			return
		}
	}

	runtimeCache := cache.GetCurrentCache()

	// check top level domain
	if tld, err := runtimeCache.GetTopLevelDomain(domainPost.TopLevelDomain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Errorf("Failed to find top level domain due to error: %s ", err.Error()).Error())
		return
	} else if tld == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Errorf("Top Level domsin %s do not exists, Please create it first ", domainPost.TopLevelDomain).Error())
		return
	}

	// check sub domain list
	if subDomains, err := runtimeCache.GetSubDomain(domainPost.TopLevelDomain, domainPost.Domain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Errorf("Failed to check existence of sub domain due to error: %s ", err.Error()).Error())
		return
	} else if subDomains != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Errorf("Sub domain %s already exists ", domainPost.Domain).Error())
		return
	}

	if err := runtimeCache.CreateOrUpdateSubDomain(domainPost); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to create sub domain due to error: %s", err.Error()))
		return
	}

	if err := notifyDns(domainPost, constants.ActionCreate); err != nil {
		log.Warnf("Failed to notify dns server due to error: %s", err.Error())
	}

	response.WriteAsJson(domainPost)
}

func UpdateSubDomain(request *restful.Request, response *restful.Response) {

	tldDomain := request.PathParameter("tld-domain")
	if len(strings.TrimSpace(tldDomain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprint("Path parameter tld-domain is not found "))
		return
	}

	domainName := request.PathParameter("domain")
	if len(strings.TrimSpace(domainName)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprint("Path parameter domain is not found "))
		return
	}

	domainPost := coredns.DNSDomain{}
	if err := request.ReadEntity(&domainPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()

	targetDomain, err := runtimeCache.GetSubDomain(tldDomain, domainName)

	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to find sub domain due to error: %s ", err.Error()))
		return
	}

	if targetDomain == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Sub domain %s not found ", domainPost.Domain))
		return
	}

	// do resolve validation
	for index, domain := range domainPost.DomainResolve {

		domainPost.DomainResolve[index].ResolveDomain = domainName
		err := utils.DnsValidate(domain)
		if err != nil {
			utils.ResponseError(response, http.StatusBadRequest, err.Error())
			return
		}
	}

	// update description and DomainResolve only
	targetDomain.Description = domainPost.Description
	targetDomain.DomainResolve = domainPost.DomainResolve

	// update sub domain
	if len(targetDomain.DomainResolve) > 0 {
		if err := runtimeCache.CreateOrUpdateSubDomain(*targetDomain); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to save sub domain to db due to error: %s ", err.Error()))
			return
		}
		if err := notifyDns(*targetDomain, constants.ActionCreate); err != nil {
			log.Warnf("Failed to notify dns server due to error: %s", err.Error())
		}
	} else {
		// remove the sub domain if no resolve record exists anymore
		log.Debug("No more resolve record exists remove the sub domain")
		if err := runtimeCache.DeleteSubDomain(tldDomain, domainName); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to remove sub domain due to error: %s ", err.Error()))
			return
		}
		if err := notifyDns(*targetDomain, constants.ActionDelete); err != nil {
			log.Warnf("Failed to notify dns server due to error: %s", err.Error())
			return
		}
	}
	response.WriteAsJson(targetDomain)
}

func notifyDns(domainPost coredns.DNSDomain, action string) error {
	runtimeCache := cache.GetCurrentCache()
	config := runtimeCache.GetServerRuntimeConfig(cache.NodeId)
	domainPost.Action = action
	// send to mq right away if config.CoreDNSConfig.NotifyDNS is set to constants.DnsNotificationModeNow
	if config.CoreDNSConfig.NotifyDNS == constants.DnsNotificationModeNow {
		dnsData, errParse := json.Marshal(&domainPost)
		if errParse != nil {
			return errParse
		}
		if err := coredns.SendToDNS(string(dnsData), config.CoreDNSConfig.Subject, cache.NodeId, config.MessageQueue); err != nil {
			return err
		}
	}
	return nil
}

func ListTopLevelDomains(request *restful.Request, response *restful.Response) {
	runtimeCache := cache.GetCurrentCache()
	result, err := runtimeCache.GetTopLevelDomainList()
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to list dns domain list due to error: %s ", err.Error()))
		return
	}
	domains := []coredns.TopLevelDomain{}
	for _, domain := range result {
		if subDomain, err := runtimeCache.GetSubDomainOfTls(domain.Domain); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to count sub domain of top level domain %s due to error: %s ", domain.Domain, err.Error()))
			return
		} else {
			domain.DomainCounts = len(subDomain)
		}
		domains = append(domains, domain)
	}
	response.WriteAsJson(domains)
}

func GetTopLevelDomain(request *restful.Request, response *restful.Response) {
	domain := request.PathParameter("domain")

	if len(strings.TrimSpace(domain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, "Path parameter domain is not found ")
		return
	}

	runtimeCache := cache.GetCurrentCache()
	result, err := runtimeCache.GetTopLevelDomain(domain)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to list top level domain due to error: %s ", err.Error()))
		return
	}

	if result == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find top level domain with name %s ", domain))
		return
	}

	if subDomain, err := runtimeCache.GetSubDomainOfTls(result.Domain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to count sub domain of top level domain %s due to error: %s ", result.Domain, err.Error()))
		return
	} else {
		result.DomainCounts = len(subDomain)
	}

	response.WriteAsJson(result)
}

func DeleteTopLevelDomain(request *restful.Request, response *restful.Response) {
	domain := request.PathParameter("domain")

	if len(strings.TrimSpace(domain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, "Path parameter domain is not found ")
		return
	}

	runtimeCache := cache.GetCurrentCache()

	if domainFound, err := runtimeCache.GetTopLevelDomain(domain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get dns domain due to error: %s ", err))
		return
	} else if domainFound == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find domain with name %s ", domain))
		return
	} else {
		if subDomain, err := runtimeCache.GetSubDomainOfTls(domain); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get sub domain of top level domain %s due to error: %s ", domain, err.Error()))
			return
		} else if len(subDomain) > 0 {
			utils.ResponseError(response, http.StatusPreconditionFailed, fmt.Sprintf("Delete sub domain first before you can delete top level domain %s ", domain))
			return
		}
	}

	err := runtimeCache.DeleteTopLevelDomain(domain)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to delete top level domain due to error: %s ", err.Error()))
		return
	}
	response.WriteHeader(http.StatusOK)
}

func UpdateTopLevelDomain(request *restful.Request, response *restful.Response) {
	domain := request.PathParameter("domain")

	if len(strings.TrimSpace(domain)) == 0 {
		utils.ResponseError(response, http.StatusInternalServerError, "Path parameter domain is not found ")
		return
	}

	domainPost := coredns.TopLevelDomain{}
	// try marsha post domain data struct
	if err := request.ReadEntity(&domainPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()
	// try to find old domain data alone with give domain id
	domainPostOld, err := runtimeCache.GetTopLevelDomain(domain)
	if err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to get dns domain due to error: %s ", err.Error()))
		return
	}

	if domainPostOld == nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Unable to find domain with name %s ", domain))
		return
	}

	// replace old domain data with new post data
	domainPostOld.Description = domainPost.Description

	if err := runtimeCache.CreateOrUpdateTopLevelDomain(*domainPostOld); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to save dns to db due to error: %s ", err.Error()))
		return
	}

	response.WriteAsJson(struct {
		Description string `json:"description,omitempty"`
	}{
		Description: domainPost.Description,
	})
}

func CreateTopLevelDomain(request *restful.Request, response *restful.Response) {
	domainPost := coredns.TopLevelDomain{}
	if err := request.ReadEntity(&domainPost); err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	// do domain validation
	err := utils.DnsValidate(domainPost)
	if err != nil {
		utils.ResponseError(response, http.StatusBadRequest, err.Error())
		return
	}

	runtimeCache := cache.GetCurrentCache()

	if existingDomain, err := runtimeCache.GetTopLevelDomain(domainPost.Domain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check existence of top level domain due to error: %s ", err.Error()))
		return
	} else if existingDomain != nil {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Top level domain %s alredy exists ", domainPost.Domain))
		return
	}

	if subDomain, err := runtimeCache.GetSubDomainOfTls(domainPost.Domain); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check existence of sub domain due to error: %s ", err.Error()))
		return
	} else if len(subDomain) > 0 {
		utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Cannot use exising sub domain %s as top level domain ", domainPost.Domain))
		return
	}

	// loop check top level domain`s root domain name is already exits
	if strings.Count(domainPost.Domain, ".") > 1 {
		if allTld, err := runtimeCache.GetTopLevelDomainList(); err != nil {
			utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to check existence of top level domain due to error: %s ", err.Error()))
			return
		} else {
			for _, existsDomain := range allTld {
				if strings.Contains(domainPost.Domain, existsDomain.Domain) {
					utils.ResponseError(response, http.StatusBadRequest, fmt.Sprintf("Top level domain %s alredy exists, cannot create %s ", existsDomain.Domain, domainPost.Domain))
					return
				}
			}
		}
	}

	if err := runtimeCache.CreateOrUpdateTopLevelDomain(domainPost); err != nil {
		utils.ResponseError(response, http.StatusInternalServerError, fmt.Sprintf("Failed to save top level domain to db due to error: %s ", err.Error()))
		return
	}

	response.WriteAsJson(&domainPost)
}
