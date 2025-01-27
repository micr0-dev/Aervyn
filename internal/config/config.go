package config

var (
	Domain      = "localhost:8080"
	Protocol    = "http"
	InstanceURL = Protocol + "://" + Domain
)

func GetActorURL(username string) string {
	return InstanceURL + "/users/" + username
}
