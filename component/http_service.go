package component

import (
	"fmt"
	"github.com/xjplke/syler/config"
	"github.com/xjplke/syler/i"
	"log"
	"net"
	"strings"
	"net/http"
	"path/filepath"
	"runtime/debug"
)

func ErrorWrap(w http.ResponseWriter) {
	if e := recover(); e != nil {
		log.Print("panic:", e, "\n", string(debug.Stack()))
		w.WriteHeader(http.StatusInternalServerError)
		if err, ok := e.(error); ok {
			w.Write([]byte(err.Error()))
		}
	}
}

func StartHttp() {
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		if handler, ok := i.ExtraAuth.(i.HttpHandler); ok {
			handler.HandleLogin(w, r)
		} else {
			BASIC_SERVICE.HandleLogin(w, r)
		}
	})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			ErrorWrap(w)
		}()
		var err error
		log.Println("handle logout!")	
		//nas := r.FormValue("nasip") //TODO
		//userip_str := r.FormValue("userip")
		
		nas := r.FormValue("nasip")
                if *config.NasIp != "" {
                        nas = *config.NasIp
                }
                userip_str := r.FormValue("userip")

                if *config.UseRemoteIpAsUserIp == true && userip_str == "" {
                        userip_str = r.Header.Get("X-Forwarded-For")
                        if userip_str != "" {
                                userip_str = strings.Split(userip_str,",")[0];
                                log.Println("Get Userip from X-Forwarded-For "+userip_str)
                        }else{
                                ip, _, _ := net.SplitHostPort(r.RemoteAddr)
                                userip_str = ip;
                        }      
                }

		log.Println("userip = "+userip_str)	
		log.Println("nasip = "+nas)	
		if userip := net.ParseIP(userip_str); userip != nil {
			if basip := net.ParseIP(nas); basip != nil {
				if _, err = Logout(userip, *config.HuaweiSecret, basip); err == nil {
					w.WriteHeader(http.StatusOK)
					return
				}
			} else {
				err = fmt.Errorf("Parse Ip err from %s", nas)
			}
		} else {
			err = fmt.Errorf("Parse Ip err from %s", userip_str)
		}
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Show login page")
		path := filepath.FromSlash(*config.LoginPage)
		http.ServeFile(w, r, path)
	})
	log.Printf("listen http on %d\n", *config.HttpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *config.HttpPort), nil)
	if err != nil {
		fmt.Println(err)
		log.Println(err)
	}
}
