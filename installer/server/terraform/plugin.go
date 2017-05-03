package terraform

import (
	"bytes"
	gtemplate "text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/coreos/terraform-provider-matchbox/matchbox"
	"github.com/hashicorp/terraform/plugin"
	"github.com/kardianos/osext"

	"github.com/coreos/tectonic-installer/installer/server/terraform/plugins/aws"
)

var plugins = map[string]*plugin.ServeOpts{
	// https://github.com/hashicorp/terraform/pull/13574 (82bda74)
	"aws": {ProviderFunc: aws.Provider},

	// https://github.com/coreos/terraform-provider-matchbox (glide)
	"matchbox": {ProviderFunc: matchbox.Provider},
}

var pluginsConfigTemplate = gtemplate.Must(gtemplate.New("").Parse(`
providers {
	{{- range $name, $plugin := .Plugins }}
	{{ $name }} = "{{ $.BinaryPath }}-TFSPACE-{{ $name }}"
	{{- end }}
}`))

// ServePlugin serves every vendored TerraForm providers/provisioners. This
// function never returns and should be the final function called in the main
// function of the plugin. Additionally, there should be no stdout/stderr
// outputs, which may interfere with the handshake and further communications.
func ServePlugin(name string) {
	p := plugins[name]
	if p == nil {
		log.Fatalf("could not find plugin %q", name)
	}

	plugin.Serve(p)
}

// BuildPluginsConfig creates a configuration that can be used by TerraForm to
// make available any provider/provisioners that are vendored in the present
// binary.
func BuildPluginsConfig() (string, error) {
	execPath, err := osext.Executable()
	if err != nil {
		return "", err
	}

	var buffer bytes.Buffer
	var data = struct {
		Plugins    map[string]*plugin.ServeOpts
		BinaryPath string
	}{plugins, execPath}

	err = pluginsConfigTemplate.Execute(&buffer, data)
	return buffer.String(), err
}
