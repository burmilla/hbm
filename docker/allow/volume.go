package allow

import (
	"fmt"

	"github.com/docker/docker/api/types/volume"
	"github.com/docker/go-plugins-helpers/authorization"
	"github.com/juliengk/go-utils"
	"github.com/juliengk/go-utils/json"
	"github.com/kassisol/hbm/docker/allow/types"
	policyobj "github.com/kassisol/hbm/object/policy"
	"github.com/kassisol/hbm/version"
	log "github.com/sirupsen/logrus"
)

func VolumeCreate(req authorization.Request, config *types.Config) *types.AllowResult {
	vol := &volume.VolumeCreateBody{}

	err := json.Decode(req.RequestBody, vol)
	if err != nil {
		return &types.AllowResult{Allow: false, Error: err.Error()}
	}

	defer utils.RecoverFunc()

	p, err := policyobj.New("sqlite", config.AppPath)
	if err != nil {
		log.WithFields(log.Fields{
			"version": version.Version,
		}).Fatal(err)
	}
	defer p.End()

	if len(vol.Driver) > 0 {
		if !p.Validate(config.Username, "volumedriver", vol.Driver, "") {
			return &types.AllowResult{
				Allow: false,
				Msg: map[string]string{
					"text":           fmt.Sprintf("Volume driver %s is not allowed", vol.Driver),
					"resource_type":  "volumedriver",
					"resource_value": vol.Driver,
				},
			}
		}
	}

	if len(vol.DriverOpts) > 0 {
		for k, v := range vol.DriverOpts {
			if vol.Driver == "local" && k == "type" && v == "tmpfs" {
				if !p.Validate(config.Username, "config", "container_create_param_tmpfs", "") {
					return &types.AllowResult{
						Allow: false,
						Msg: map[string]string{
							"text":           "--tmpfs param is not allowed",
							"resource_type":  "config",
							"resource_value": "container_create_param_tmpfs",
						},
					}
				}
			}
		}
	}

	return &types.AllowResult{Allow: true}
}
