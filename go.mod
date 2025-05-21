module web-workshop-eval3

go 1.21

require (
	github.com/google/uuid v1.6.0
	golang.org/x/crypto v0.21.0
)

replace (
	github.com/yourusername/webworkshop-eva3/web/modules/producto => ./web/modules/producto
	github.com/yourusername/webworkshop-eva3/web/modules/usuario => ./web/modules/usuario
	github.com/yourusername/webworkshop-eva3/web/public => ./web/public
)
