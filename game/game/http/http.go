package http

import (
	"fmt"
	"net/http"
	"simple_game/game/config"
	"simple_game/game/pkg"
)

func StartHttp() {
	// http 路由
	http.HandleFunc("/", index)
	http.HandleFunc("/hello", index)

	addrCfg, port := config.Conf.Http.Addr, config.Conf.Http.Port
	addr := fmt.Sprintf("%s:%s", addrCfg, port)
	pkg.INFO("simple game http start suc!")
	if err := http.ListenAndServe(addr, nil); err != nil {
		pkg.ERROR("simple game http start fail", err)
		return
	}
}

func index(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "hello simple game")
	pkg.INFO("remote:", request.RemoteAddr)
}
