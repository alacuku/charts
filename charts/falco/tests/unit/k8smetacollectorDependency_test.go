package unit

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

const chartPath = "../../"

// Using the default values we want to test that all the expected resources for the k8s-metacollector are rendered.
func TestRenderedResourcesWithDefaultValues(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	releaseName := "rendered-resources"

	// Template files that we expect to be rendered.
	templateFiles := []string{
		"clusterrole.yaml",
		"clusterrolebinding.yaml",
		"deployment.yaml",
		"service.yaml",
		"serviceaccount.yaml",
	}

	require.NoError(t, err)

	options := &helm.Options{}

	// Template the chart using the default values.yaml file.
	output, err := helm.RenderTemplateE(t, options, helmChartPath, releaseName, nil)
	require.NoError(t, err)

	// Extract all rendered files from the output.
	pattern := `# Source: falco/charts/k8s-metacollector/templates/([^\n]+)`
	re := regexp.MustCompile(pattern)
	matches := re.FindAllStringSubmatch(output, -1)

	var renderedTemplates []string
	for _, match := range matches {
		// Filter out test templates.
		if !strings.Contains(match[1], "test-") {
			renderedTemplates = append(renderedTemplates, match[1])
		}
	}

	// Assert that the rendered resources are equal tho the expected ones.
	require.Equal(t, len(renderedTemplates), len(templateFiles), "should be equal")

	for _, rendered := range renderedTemplates {
		require.True(t, slices.Contains(templateFiles, rendered), "template files should contain all the rendered files")
	}
}

func TestPluginConfigurationInFalcoConfig(t *testing.T) {
	t.Parallel()

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	releaseName := "rendered-resources"

	testCases := []struct {
		name     string
		values   map[string]string
		expected func(t *testing.T, config any)
	}{
		{
			"defaultValues",
			nil,
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, fmt.Sprintf("%s-k8s-metacollector.default.svc", releaseName), hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},
		{
			"overrideK8s-metacollectorNamespace",
			map[string]string{
				"k8s-metacollector.namespaceOverride": "test",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, fmt.Sprintf("%s-k8s-metacollector.test.svc", releaseName), hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},
		{
			"overrideK8s-metacollectorName",
			map[string]string{
				"k8s-metacollector.fullnameOverride": "collector",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, "collector.default.svc", hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},

		{
			"overrideK8s-metacollectorNamespaceAndName",
			map[string]string{
				"k8s-metacollector.namespaceOverride": "test",
				"k8s-metacollector.fullnameOverride":  "collector",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, "collector.test.svc", hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},
		{
			"set CollectorHostname",
			map[string]string{
				"collectors.kubernetes.collectorHostname": "test",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, "test", hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},

		{
			"set CollectorHostname and namespace name",
			map[string]string{
				"collectors.kubernetes.collectorHostname": "test-with-override",
				"k8s-metacollector.namespaceOverride":     "test",
				"k8s-metacollector.fullnameOverride":      "collector",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(45000), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, "test-with-override", hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},

		{
			"set collectorPort",
			map[string]string{
				"collectors.kubernetes.collectorPort": "8888",
			},
			func(t *testing.T, config any) {
				plugin := config.(map[string]interface{})
				// Get init config.
				initConfig, ok := plugin["init_config"]
				require.True(t, ok)
				initConfigMap := initConfig.(map[string]interface{})
				// Check that the collector port is correctly set.
				port := initConfigMap["collectorPort"]
				require.Equal(t, float64(8888), port.(float64))
				// Check that the collector nodeName is correctly set.
				nodeName := initConfigMap["nodename"]
				require.Equal(t, "${FALCO_K8S_NODE_NAME}", nodeName.(string))
				// Check that the collector hostname is correctly set.
				hostName := initConfigMap["collectorHostname"]
				require.Equal(t, fmt.Sprintf("%s-k8s-metacollector.default.svc", releaseName), hostName.(string))

				// Check that the library path is set.
				libPath := plugin["library_path"]
				require.Equal(t, "libk8smeta.so", libPath)
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := &helm.Options{SetValues: testCase.values}
			output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/configmap.yaml"})

			var cm corev1.ConfigMap
			helm.UnmarshalK8SYaml(t, output, &cm)
			var config map[string]interface{}

			helm.UnmarshalK8SYaml(t, cm.Data["falco.yaml"], &config)
			plugins := config["plugins"]
			pluginsArray := plugins.([]interface{})
			// Find the k8smeta plugin configuration.
			for _, plugin := range pluginsArray {
				if name, ok := plugin.(map[string]interface{})["name"]; ok && name == "k8smeta" {
					testCase.expected(t, plugin)
				}
			}
			// Check that the plugin has been added to the ones that need to be loaded.
			loadplugins := config["load_plugins"]
			require.True(t, slices.Contains(loadplugins.([]interface{}), "k8smeta"))
		})
	}
}

// Test that the helper does not overwrite user's configuration.
func TestPluginConfigurationUniqueEntries(t *testing.T) {
	t.Parallel()

	pluginsJSON := `[
    {
        "init_config": null,
        "library_path": "libk8saudit.so",
        "name": "k8saudit",
        "open_params": "http://:9765/k8s-audit"
    },
    {
        "library_path": "libcloudtrail.so",
        "name": "cloudtrail"
    },
    {
        "init_config": "",
        "library_path": "libjson.so",
        "name": "json"
    },
    {
        "init_config": {
            "collectorHostname": "rendered-resources-k8s-metacollector.default.svc",
            "collectorPort": 45000,
            "nodename": "${FALCO_K8S_NODE_NAME}"
        },
        "library_path": "libk8smeta.so",
        "name": "k8smeta"
    }
]`

	loadPluginsJSON := `[
    "k8smeta",
	"k8saudit"
]`
	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	releaseName := "rendered-resources"

	options := &helm.Options{SetJsonValues: map[string]string{
		"falco.plugins":      pluginsJSON,
		"falco.load_plugins": loadPluginsJSON,
	}}
	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/configmap.yaml"})

	var cm corev1.ConfigMap
	helm.UnmarshalK8SYaml(t, output, &cm)
	var config map[string]interface{}

	helm.UnmarshalK8SYaml(t, cm.Data["falco.yaml"], &config)
	plugins := config["plugins"]

	out, err := json.MarshalIndent(plugins, "", "    ")
	require.NoError(t, err)
	require.Equal(t, pluginsJSON, string(out))
	pluginsArray := plugins.([]interface{})
	// Find the k8smeta plugin configuration.
	numConfigK8smeta := 0
	for _, plugin := range pluginsArray {
		if name, ok := plugin.(map[string]interface{})["name"]; ok && name == "k8smeta" {
			numConfigK8smeta++
		}
	}

	require.Equal(t, 1, numConfigK8smeta)

	// Check that the plugin has been added to the ones that need to be loaded.
	loadplugins := config["load_plugins"]
	require.Len(t, loadplugins.([]interface{}), 2)
	require.True(t, slices.Contains(loadplugins.([]interface{}), "k8smeta"))
}

// Test that the helper does not overwrite user's configuration.
func TestFalcoctlRefs(t *testing.T) {
	t.Parallel()

	pluginsJSON := `[
	    {
	        "init_config": null,
	        "library_path": "libk8saudit.so",
	        "name": "k8saudit",
	        "open_params": "http://:9765/k8s-audit"
	    },
	    {
	        "library_path": "libcloudtrail.so",
	        "name": "cloudtrail"
	    },
	    {
	        "init_config": "",
	        "library_path": "libjson.so",
	        "name": "json"
	    },
	    {
	        "init_config": {
	            "collectorHostname": "rendered-resources-k8s-metacollector.default.svc",
	            "collectorPort": 45000,
	            "nodename": "${FALCO_K8S_NODE_NAME}"
	        },
	        "library_path": "libk8smeta.so",
	        "name": "k8smeta"
	    }
	]`

	testFunc := func(t *testing.T, config any) {
		// Get artifact configuration map.
		configMap := config.(map[string]interface{})
		artifactConfig := (configMap["artifact"]).(map[string]interface{})
		// Test allowed types.
		allowedTypes := artifactConfig["allowedTypes"]
		require.Len(t, allowedTypes, 2)
		require.True(t, slices.Contains(allowedTypes.([]interface{}), "plugin"))
		require.True(t, slices.Contains(allowedTypes.([]interface{}), "rulesfile"))
		// Test plugin reference.
		refs := artifactConfig["install"].(map[string]interface{})["refs"].([]interface{})
		require.Len(t, refs, 2)
		require.True(t, slices.Contains(refs, "falco-rules:2"))
		require.True(t, slices.Contains(refs, "ghcr.io/falcosecurity/plugins/plugin/k8smeta:0.1.0-alpha"))
	}

	testCases := []struct {
		name       string
		valuesJSON map[string]string
		expected   func(t *testing.T, config any)
	}{
		{
			"defaultValues",
			nil,
			testFunc,
		},
		{
			"setPluginConfiguration",
			map[string]string{
				"falco.plugins": pluginsJSON,
			},
			testFunc,
		},
	}

	helmChartPath, err := filepath.Abs(chartPath)
	require.NoError(t, err)

	releaseName := "rendered-resources"

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			options := &helm.Options{SetJsonValues: testCase.valuesJSON}
			output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/falcoctl-configmap.yaml"})

			var cm corev1.ConfigMap
			helm.UnmarshalK8SYaml(t, output, &cm)
			var config map[string]interface{}
			helm.UnmarshalK8SYaml(t, cm.Data["falcoctl.yaml"], &config)
			testCase.expected(t, config)
		})
	}
}
