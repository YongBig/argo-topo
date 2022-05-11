package main

import (
	scontext "context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/YongBig/argo-topo/pkg/Config"
	"github.com/YongBig/argo-topo/pkg/DB"
	argo "github.com/YongBig/argo-topo/pkg/controllers"
	"github.com/YongBig/argo-topo/pkg/response"
	"github.com/kataras/iris/v12"
	"k8s.io/client-go/rest"
	log "k8s.io/klog/v2"
	ctrlsignals "sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

var ArgoController *argo.Controller

func main() {
	ctx := ctrlsignals.SetupSignalHandler()
	ctx, cancel := scontext.WithCancel(ctx)
	Cfg := GetConfig()
	//note:  create manager && argo engine cache
	config := Config.NewConfig(Cfg)
	go Run(ctx, cancel, config)

	// api server
	app := iris.New()
	// Api get responsetree
	app.Get("/topologys/{namespace:string}/{name:string}", func(ctx iris.Context) {
		name := ctx.Params().Get("name")
		namespace := ctx.Params().Get("namespace")
		//note: 获取redis数据
		result, ok, err := DB.RDB.Get(fmt.Sprintf("%s/%s", namespace, name))
		if ok && err != nil {
			response.ErrResponse(ctx, err)
			return
		}
		if !ok {
			app, err := ArgoController.GetApp(scontext.Background(), name, namespace)

			if err != nil {
				response.ErrResponse(ctx, err)
				return
			}
			if app == nil {
				response.NotFoundResponse(ctx, fmt.Errorf("%s/%s is not found in k8s", name, namespace))
				return
			}
			appTopology, buildErr := ArgoController.Helper.BuildResourceTreeForApp(app)
			if buildErr != nil {
				response.ErrResponse(ctx, err)
				return
			}
			cacheKey := ArgoController.Helper.CacheKey(app)
			err = DB.RDB.Set(cacheKey, appTopology.Marshal())
			if err != nil {
				response.ErrResponse(ctx, err)
				return
			}
			response.SuccessResponse(ctx, appTopology)
			return
		}
		var relStruct map[string]interface{}
		err = json.Unmarshal([]byte(result), &relStruct)
		if err != nil {
			response.ErrResponse(ctx, err)
			return
		}
		response.SuccessResponse(ctx, relStruct)
		return
	})

	err := app.Run(iris.Addr("0.0.0.0:8081"))
	if err != nil {
		panic(err)
	}
}

func Run(ctx scontext.Context, cancel scontext.CancelFunc, config *Config.Config) {
	cfg := argo.CreateConfig()
	k8sclient := config.Manager.GetClient()
	arogoController := argo.NewArgoAppController(cfg, k8sclient, config.ArgoEngineClusterCache)
	ArgoController = arogoController
	go func() {
		err := runControllers(config, ctx, cancel, arogoController)
		if err != nil {
			panic(err)
		}
	}()
}

func runControllers(cfg *Config.Config, context scontext.Context, cancel scontext.CancelFunc, controller *argo.Controller) error {
	defer cancel()
	log.Info("start controller manager")
	err := controller.SetupManager(cfg.Manager)
	if err != nil {
		return err
	}
	log.Info("Start manager")
	return cfg.Manager.Start(context)
}

func GetConfig() *rest.Config {
	certData, _ := base64.StdEncoding.DecodeString("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJSWhHSVRPZzgvMk13RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB5TVRFeE1EZ3dOakU0TVRSYUZ3MHlNakV4TURnd05qRTVORGRhTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXFySms4V2R3NU1DcERudXAKaXNOclhvTXpTanJjZmlBckxhNWVPUVpBUlFZaHZ6R3hvRVdMem1vb0hWdGxqZ01GaHhOZU5lOVhiTlY0bmdwWgp2S1ZEUiszNVBlUjJvQUtudGtKVWc3VWVkZ01TZGxCd0lxWHB1Nmo3aURidVZMY2NXMklhSHZvcHIzMW94RTJwCkhnWVpma29mK25JdzZzTWM5ZHlDUjF2cjNqVG53SnJkc1pTMEk5cUNhOEczeDlWNElZd2Z1eXhIaWp0NzRoM3oKc3NNY09iY0Y2bkJqNDNEOVhISWRMcEtyS2J2WFVLS2NkU2E4WUlWZUFHRWkrYzc4RXZpWEd0SlhpVHd4TnlLNwpvRnA3azcyazN0V0dvdzV4eHUyZStRYzBxMzkzWVlsOXZONFVkMFpyWkdua2NsTjlVV2wvWHJFdkNSenh4a0w2CkFjRVBkd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDb1NZdEZRaTRIb3lIazd1RUQwWUVFczJqOFQxMFh1TDNBZgpFTGwrc243NnVnOTZTSjBBR2lnVkFWUG5zWEhPeVZPc3pYMnUxQThYRmxTR1Y5YU5IdUxvbWVFYnBCWDUvdkVuClZIc3R5TjZmNXgwS0VhcGFvdHZXenN0aXRPVGxUU0hnY1lvZFVHdzJqZDIrdFgvU0lKMHY2RzVPK0VsQ05LMGsKYW9ncVVTUlV1c0dYY1FSZHZqbGh2UUNyUWZtZkQwK1piV3dPQ0x2ZnFaSkJIS00zM3U2SzJwMjF5cG5QcFU1OApFanU2cUliMEZDTytEdVBXWnl6cmpCQVBhNVdpUWxVRmNPV3g0b2RFTWc0SUVxSHZXcGplQlBOQW9EeXh1cVl1Cm1yeEgvM2ZKV1NEY3FVZXg4am5UNVhxajI4RjZhQUpHVHdqM2phVGduRjhqbytZWmtkVT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=")
	keyData, _ := base64.StdEncoding.DecodeString("LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBcXJKazhXZHc1TUNwRG51cGlzTnJYb016U2pyY2ZpQXJMYTVlT1FaQVJRWWh2ekd4Cm9FV0x6bW9vSFZ0bGpnTUZoeE5lTmU5WGJOVjRuZ3BadktWRFIrMzVQZVIyb0FLbnRrSlVnN1VlZGdNU2RsQncKSXFYcHU2ajdpRGJ1VkxjY1cySWFIdm9wcjMxb3hFMnBIZ1laZmtvZituSXc2c01jOWR5Q1IxdnIzalRud0pyZApzWlMwSTlxQ2E4RzN4OVY0SVl3ZnV5eEhpanQ3NGgzenNzTWNPYmNGNm5CajQzRDlYSElkTHBLcktidlhVS0tjCmRTYThZSVZlQUdFaStjNzhFdmlYR3RKWGlUd3hOeUs3b0ZwN2s3MmszdFdHb3c1eHh1MmUrUWMwcTM5M1lZbDkKdk40VWQwWnJaR25rY2xOOVVXbC9YckV2Q1J6eHhrTDZBY0VQZHdJREFRQUJBb0lCQUJ2UkpwSVFnVjFGNkVicgo4QjBrNjRKamJ5QlJwMDBHZ1FMWXY1SWJhcTNVNmZyMlpqUHdJWEJwN0UrY2JWaFBOYjlsY0p6cGZCM0lTL2UxClNCcHQ5Q0RzcndsZkNkWFptT3NpdEJNaW1Pd3lZL2ZUOC9JeGgzMkZkRGdtZTRCaXRzUk5vR1FiOEY4ZzJNbmsKdmdLZWk5a0F5MlZNNXB2YllBVFJBb29PZU1tbElWaDlBN1BMY0VxdkZxMEtoTXdPU1FENlBuamhrVTBUOWltbwovbTVMdlRpMkdtTklRT1ozcGJtQ3pLaU1SbXY4OWFXREFaOEdINW1sSyt5R0RKZXhFbkQ2Q0h5dEl6K1JSUUl0CjNhN2YxdVFYSjR0VHRWWE5ORVJRazZrSW0wb25PbkVZd1YzM1ZpODFaZGFYaUJYQzg1blFLbzdlMjR1SG9kSzkKTCsvdzY4RUNnWUVBekNZNzRCaG04K2RpWmZyMmhUeHJqMHY3VGhFY1ozWE52QlF4YXdBYXNFYzN5Q3BLenNIUgpIK3M0VGNKYzFUR3VXZ1loOEI4WkZycThlNXhnbDVvM2xvMUxCVkw1ajNNQzhEeWMyeFgxUGFLQkFDb0E1WnNwCjZiR25wVEE5UVBkUXQyWkcxODJUdUJxZmxlY2R4VStPcFM3Vk5nUzIzZTBWRlkwTC9ScGc0N3NDZ1lFQTFnMFYKUVJRZUh0R3NiZGI1TFFTSVJ6S3U4RVZqREFaKzdoTmNoUkpMVldlVk9DZ0VmK0Jwc0NpNlZXSUdUZEFWNHorTwpnQzYvb2pPU3JLZzR0QlNyU2R3Ti9XT2M5RVZKMDhqSnczaWZBKysvNlgwa1lxd3JvQVJJUndadTlEYUU3cEpLCmM0aE1YdVBCd0V6WVZJS1RoM3RHSXFPVWJ5aWNXSjU4NnBCTHdYVUNnWUVBaTNHdVFsTEl0OThid2liYkRvVUgKdnppYUxtZlhxLysyaUxxT1N0VW1aYlF2c1FUYVZrSGpNMWM2L1RvK3FNMG5sNHhLMERhZHIzM2IwdDhzeDBEcQpxV1pYa1FwdE5vUEx4UWJSNllBbEpIV0VnZlV1NmFiRHlVRzFEa3RWKzdNeXFpTXRUcWk0TnUvUWc5YjY2ZFIrCnpldWdiU1pwTmt1RHRGWEVrNXphQTVNQ2dZRUFrK1ZpUkI4RVdNTUM0cm5nWFF4K3BNTU9RSkdReUNSTTIyNmgKUklqSmFHOHptU045U0dYa1lJVWppZzg2ejlUdzZwMWxkb2ZXZk5vcGhBYVBkMDI0dEVYSm5NU1JFKzR6L3BNRApaWDRZVVAzOG1mV1BpR1h4bHBTZTVBUTc4WjBoNkQxSUY5K2E5UTFsTjl0Z3RiT3EvN2RiVkYrMkZiLzNsdnVhCnorOTNpR2tDZ1lCand5QzJSRjhyZE9WUFYvZDJhYWJIRkk4RzNkRmFvczNBektoaXg4OWlwd2MwMGt3MHFpeGEKNDlRNnhWL1A3eko0alJlRzFrYVJOeHRLcFRSZUdDejBkZGJrRjFYb21zTXArdkRLOFltWGdjZDhqeWFxb0RYTwpibVlyTlZiajl3VmtLVkxuZ2RBSGY1RHo0ZksyY3ZhUG9zbmhxM29IK3FScUpCcG1iTExUV0E9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=")
	cAData, _ := base64.StdEncoding.DecodeString("LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFWTVJNd0VRWURWUVFERXdwcmRXSmwKY201bGRHVnpNQjRYRFRJeE1URXdPREEyTVRneE5Gb1hEVE14TVRFd05qQTJNVGd4TkZvd0ZURVRNQkVHQTFVRQpBeE1LYTNWaVpYSnVaWFJsY3pDQ0FTSXdEUVlKS29aSWh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTUJECnp1MXlMbVZDRUs1eTdMUzV5ZzJyVThmYndBVEExWjQ0Ly9LaGhTSzAvblFmbzlSOUx3NE1YdS9PblIwaDYzK3QKWGdpRG1FZndwR0hCbEhSTHFSbFR5QmlTMFcrTjNOKzhuRHg5U1pjQUxTTi9xK2lBLzBBOHhyOHl1Rnd5Tnd1eQpFdjg0SkVFaEJZMUJpS0ZIY3JsSkZNRS80WlpURVRwUWowSkQxS3AvL0pRTTBmMjMwSVJXbDVnMFRaM2FJbVp4CnVBbzFFN3FvTkZObDhhKzZyRkp1eDZuaFlqYkJROThmNGIxRzRLV1VnSFNSNHVIVWpQdEczSTdNdHNhcUxLMjMKMFNPNVlBekRqSVZzSWtSbVhsUCt5L3MvMWl0c0hacGJ3VXhvTGc3VXpBR2NNK0tsWWgyNVRuWEt0UjVUK3VLdgpOdGluZHdNYWk2aUViczhGK3NNQ0F3RUFBYU1qTUNFd0RnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCCi93UUZNQU1CQWY4d0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFDZlFObWE1ODlLcHBBRCtoSlRBK2lJZ3drSzYKeWRDMGRhbnVSYzJodlp4NjlpYTZ3dko2Tll2NklzMzdmYm1aVXp1TEp0bGJvcFA3UnVjZUo0MzQrNDV2OWV0QQp6Y2doUUdUSW9uNDViR0xVTWRXN0Y0R2YzSGlYTTZFaDNHRUxmL3hqYVBEM1pRYTU2R3l2VEVZWnlBV3V4anNWClRSOUpTZUlLQlVGUFJHem9qcHdOSFJSNFdEYUJBMWNKUkV6TUk0dW9mejY2WFVUY3pLb3NlYU84MTVmNEtHOE4KTUJNYjNjbGpGZDk2eG83cUExYUduZ3VuRmErSDNGMUhWVUZMSnVNKzNsdE45UXBhdmdWY0ZIL2hTSGRicTdnNQpRcGlxN3V0NEI0K3JaTnFHdzBGSkJCK3VTT1cxSkoyOE9EL29ESXlIMGMxNVUvbFRmQk5DbWt0L01HQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=")
	var Cfg = &rest.Config{
		Host: "https://192.168.2.141:6443",
		TLSClientConfig: rest.TLSClientConfig{
			CertData: certData,
			KeyData:  keyData,
			CAData:   cAData,
		},
	}
	return Cfg
}
