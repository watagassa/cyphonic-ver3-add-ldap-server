# !bin/bash

# Copy cyphonic.lua
dest_dir="/usr/share/wireshark"
cyphonic_lua="cyphonic.lua"

if [ -f "$cyphonic_lua" ]; then
  sudo cp "$cyphonic_lua" "$dest_dir"
else
  echo "cyphonic.lua not found."
  exit 1
fi

# Modify init.lua
dest_dir="/etc/wireshark"
init_lua="$dest_dir/init.lua"

if [ -f "$init_lua" ]; then
  if grep -q "enable_lua =" "$init_lua"; then
    sudo sed -i 's/enable_lua =.*/enable_lua = true/' "$init_lua"
  elif grep -q "disable_lua =" "$init_lua"; then
    sudo sed -i 's/disable_lua =.*/disable_lua = false/' "$init_lua"
  else
    sudo sed -i '1i enable_lua = true' "$init_lua"
    sudo sed -i '1i disable_lua = false' "$init_lua"
  fi

  if ! grep -q "dofile(DATA_DIR..\"cyphonic.lua\")" "$init_lua"; then
    sudo sed -i '$ a\\ndofile(DATA_DIR.."cyphonic.lua")' "$init_lua"
  fi
else
  echo "init.lua not found."
  exit 1
fi
