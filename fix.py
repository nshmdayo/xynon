import sys
import re

with open('examples/plugins/rate-limit/main.go', 'r') as f:
    content = f.read()

content = content.replace('"strconv"', '"strconv"\n\t"encoding/json"')

old_block = """	if len(cfgData) > 0 {
		s := string(cfgData)
		// very basic manual json parsing
		if idx := strings.Index(s, "\\"key_header\\""); idx != -1 {
			if start := strings.Index(s[idx:], ":"); start != -1 {
				valStart := strings.Index(s[idx+start:], "\\"")
				if valStart != -1 {
					valEnd := strings.Index(s[idx+start+valStart+1:], "\\"")
					if valEnd != -1 {
						cfg.KeyHeader = s[idx+start+valStart+1 : idx+start+valStart+1+valEnd]
					}
				}
			}
		}
		if idx := strings.Index(s, "\\"capacity\\""); idx != -1 {
			if start := strings.Index(s[idx:], ":"); start != -1 {
				// parse number
				numStr := ""
				for i := idx + start + 1; i < len(s); i++ {
					if s[i] >= '0' && s[i] <= '9' {
						numStr += string(s[i])
					} else if s[i] != ' ' && s[i] != '\\n' && s[i] != '\\r' && len(numStr) > 0 {
						break
					}
				}
				if val, err := strconv.Atoi(numStr); err == nil {
					cfg.Capacity = val
				}
			}
		}
		if idx := strings.Index(s, "\\"refill_rate\\""); idx != -1 {
			if start := strings.Index(s[idx:], ":"); start != -1 {
				numStr := ""
				for i := idx + start + 1; i < len(s); i++ {
					if s[i] >= '0' && s[i] <= '9' {
						numStr += string(s[i])
					} else if s[i] != ' ' && s[i] != '\\n' && s[i] != '\\r' && len(numStr) > 0 {
						break
					}
				}
				if val, err := strconv.Atoi(numStr); err == nil {
					cfg.RefillRate = val
				}
			}
		}
	}"""

new_block = """	if len(cfgData) > 0 {
		_ = json.Unmarshal(cfgData, &cfg)
	}"""

content = content.replace(old_block, new_block)

with open('examples/plugins/rate-limit/main.go', 'w') as f:
    f.write(content)
