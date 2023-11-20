package main

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

type YamlConfig struct {
	Software ConfigSoftware `yaml:"software"`
}

type ConfigSoftware struct {
	SoftwareType string `yaml:"type"`
	JavaVersion  string `yaml:"java_version,omitempty"`
}

func ConvertConfig() {

	fmt.Println("\033[1m\033[38;5;220mConverting Old Config, this may take a bit...\033[0m")

	files, err := os.ReadDir(fmt.Sprintf("%s/", os.Getenv("HOME")))
	if err != nil {
		fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
		return
	}

	var newConfig YamlConfig

	for _, file := range files {

		switch file.Name() {
		case "velocity.toml":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_proxy_velocity",
					JavaVersion:  "17.0.7-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "waterfall.yml":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_proxy_waterfall",
					JavaVersion:  "17.0.7-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "config.yml":
			cFile, err := os.Open(file.Name())
			if err != nil {
				fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
				return
			}

			defer cFile.Close()

			reader := bufio.NewReader(cFile)

			for {
				line, err := reader.ReadString('\n')
				if err != nil {
					if err.Error() == "EOF" {
						break
					}
					fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
					return
				}

				if strings.Contains(line, "bungeecord") {
					newConfig = YamlConfig{
						Software: ConfigSoftware{
							SoftwareType: "mc_proxy_waterfall",
							JavaVersion:  "17.0.7-tem",
						},
					}
					break
				}
			}

		case "insurgency.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "steam_insurgency",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "njsbot.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "discord_nodejs",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "nodemonnjsbot.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "discord_nodejsnodemon",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "pybot.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "discord_python",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "phpbot.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "discord_php",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "java8":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					JavaVersion: "8.0.382-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "java11":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					JavaVersion: "11.0.20-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "java16":
		case "java17":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					JavaVersion: "17.0.8-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "bedrock_server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_bedrock_vanilla",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "PocketMine-MP.phar":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_bedrock_pmmp",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "fabric-server-launch.jar":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_java_fabric",
					JavaVersion:  "17.0.8-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "unix_args.txt":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_java_forge",
					JavaVersion:  "17.0.8-tem",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "Cuberite.server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_java_cuberite",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "magma.yml":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "mc_java_magma",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "Lavalink.jar":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "voice_lavalink",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "TeaSpeakServer":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "voice_teaspeak",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		case "ts3server":
			newConfig = YamlConfig{
				Software: ConfigSoftware{
					SoftwareType: "voice_teamspeak3",
				},
			}
			os.Remove(fmt.Sprintf("%s/%s", os.Getenv("HOME"), file.Name()))
		}

	}

	if !reflect.DeepEqual(newConfig, reflect.Zero(reflect.TypeOf(newConfig)).Interface()) {

		data, err := yaml.Marshal(&newConfig)
		if err != nil {
			fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
			return
		}

		file, err := os.Create(fmt.Sprintf("%s/eggactyl_config.yml", os.Getenv("HOME")))
		if err != nil {
			fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
			return
		}

		defer file.Close()

		writer := bufio.NewWriter(file)

		_, err = writer.Write(data)
		if err != nil {
			fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
			return
		}

		err = writer.Flush()
		if err != nil {
			fmt.Println("\033[1m\033[38;5;210mUh Oh! I couldn't convert the config... You can use the \"Convert Old Config\" option to remake the config \033[0m")
			return
		}

	}

}
