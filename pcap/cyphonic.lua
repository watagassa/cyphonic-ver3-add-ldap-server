do
	-- Protocol Definition
	cyphonic_proto = Proto("cyphonic", "CYPHONIC Protocol")

	-- Dissector Function
	cyphonic_proto.dissector = function(buffer, pinfo, tree)
		-- Type
		local type_range = buffer(6, 1)
		local type = type_range:uint()

		-- Message Length
		local msg_len_range = buffer(12, 2)
		local msg_len = msg_len_range:uint()

		-- Protocol Setting
		set_protocol(buffer, pinfo)

		local subtree = tree:add("CYPHONIC Protocol")

		-- Set the long name of the Type
		type_name = type_longname()

		-- Base Header
		dispatch(buffer(0, 32):tvb(), subtree, type_name)

		-- Processing by Type (After Base Header)
		dispatch2(type, buffer(32, msg_len - 32):tvb(), subtree, type_name)
	end

	-- Load udp.port table
	udp_table = DissectorTable.get("udp.port")

	-- Port number Setting (node to nms)
	udp_table:add(4502, cyphonic_proto)

	-- Port number Setting (node to node)
	udp_table:add(30000, cyphonic_proto)
end

-- Protocol Settings
function set_protocol(buffer, pinfo)
	-- Type
	local type = buffer(6, 1):uint()

	if type == 0 then
		pinfo.cols.protocol = "CYPHONIC(Def)"
	elseif type == 1 then
		pinfo.cols.protocol = "CYPHONIC(KS)"
	elseif type == 2 then
		pinfo.cols.protocol = "CYPHONIC(Ack)"
	elseif type == 3 then
		pinfo.cols.protocol = "CYPHONIC(LReq)"
	elseif type == 4 then
		pinfo.cols.protocol = "CYPHONIC(LRes)"
	elseif type == 5 then
		pinfo.cols.protocol = "CYPHONIC(KD)"
	elseif type == 6 then
		pinfo.cols.protocol = "CYPHONIC(RegReq)"
	elseif type == 7 then
		pinfo.cols.protocol = "CYPHONIC(RegRes)"
	elseif type == 8 then
		pinfo.cols.protocol = "CYPHONIC(KA)"
	elseif type == 9 then
		pinfo.cols.protocol = "CYPHONIC(ND)"
	elseif type == 10 then
		pinfo.cols.protocol = "CYPHONIC(NR)"
	elseif type == 11 then
		pinfo.cols.protocol = "CYPHONIC(DReq)"
	elseif type == 12 then
		pinfo.cols.protocol = "CYPHONIC(RDirCN)"
	elseif type == 13 then
		pinfo.cols.protocol = "CYPHONIC(RDirConf)"
	elseif type == 14 then
		pinfo.cols.protocol = "CYPHONIC(RDirMN)"
	elseif type == 15 then
		pinfo.cols.protocol = "CYPHONIC(NIFReq)"
	elseif type == 16 then
		pinfo.cols.protocol = "CYPHONIC(NIFRes)"
	elseif type == 17 then
		pinfo.cols.protocol = "CYPHONIC(RelayReq)"
	elseif type == 18 then
		pinfo.cols.protocol = "CYPHONIC(RelayReq)"
	elseif type == 19 then
		pinfo.cols.protocol = "CYPHONIC(NReq)"
	elseif type == 20 then
		pinfo.cols.protocol = "CYPHONIC(RouteReq)"
	elseif type == 21 then
		pinfo.cols.protocol = "CYPHONIC(HP)"
	elseif type == 22 then
		pinfo.cols.protocol = "CYPHONIC(TReq)"
	elseif type == 23 then
		pinfo.cols.protocol = "CYPHONIC(TRes)"
	elseif type == 24 then
		pinfo.cols.protocol = "CYPHONIC(CapM)"
	elseif type == 25 then
		pinfo.cols.protocol = "CYPHONIC(KAack)"
	elseif type == 26 then
		pinfo.cols.protocol = "CYPHONIC(MCReq)"
	else
		pinfo.cols.protocol = "CYPHONIC(etc)"
	end
end

-- Set the long name of the Type
function type_longname()
	-- Type
	local type_name = {
		[0] = "Default",
		[1] = "Key Sharing",
		[2] = "Ack",
		[3] = "Login Request",
		[4] = "Login Response",
		[5] = "Key Distribution",
		[6] = "Registration Request",
		[7] = "Registration Response",
		[8] = "Keep Alive",
		[9] = "Notification Direction",
		[10] = "Notification Registration",
		[11] = "Direction Request",
		[12] = "Route Direction To CN",
		[13] = "Route Direction Confirmation",
		[14] = "Route Direction To MN",
		[15] = "Node Information Request",
		[16] = "Node Information Response",
		[17] = "Relay Request",
		[18] = "Relay Response",
		[19] = "Notification Request",
		[20] = "Route Request",
		[21] = "Hole Punching",
		[22] = "Tunnel Request",
		[23] = "Tunnel Response",
		[24] = "Capsule Message",
		[25] = "Keep Alive Ack",
		[26] = "Multi Cast Request",
	}

	return type_name
end

-- Base Header
function dispatch(buffer, tree, type_longname)
	-- Flag
	local flags = {
		[0] = "Default",
		[1] = "Unclear Error Unknown",
		[2] = "No Such Name",
		[3] = "NG FQDN",
		[4] = "DB No Record",
		[5] = "No Match Password",
		[6] = "Relay TT Update Failed",
		[7] = "Table No Entry",
		[8] = "NodeType MN",
		[9] = "NodeType CN",
		[10] = "First Registration",
		[11] = "Use Virtual IPv4",
		[12] = "Optimization",
	}

	-- Next Opt (Need to add value)
	local opt_no = {
		[0] = "",
		[1] = "",
		[2] = "",
		[3] = "",
		[4] = "",
		[5] = "",
		[6] = "",
	}

	-- Header
	local subtree = tree:add(buffer(0), "CYPHONIC Header", buffer(0):tvb())

	subtree:add(buffer(0, 4), "Transaction ID: " .. buffer(0, 4):uint())
	subtree:add(buffer(4, 1), "Version: " .. buffer(4, 1):uint())
	subtree:add(buffer(5, 1), "Flag: " .. buffer(5, 1):uint(), "(" .. flags[buffer(5, 1):uint()] .. ")")
	subtree:add(buffer(6, 1), "Type: " .. buffer(6, 1):uint(), "(" .. type_longname[buffer(6, 1):uint()] .. ")")
	subtree:add(buffer(7, 1), "Count: " .. buffer(7, 1):uint())
	subtree:add(buffer(8, 4), "Sequence Number: " .. buffer(8, 4):uint())
	subtree:add(buffer(12, 2), "Message Length: " .. buffer(12, 2):uint())
	subtree:add(buffer(14, 1), "Next Opt: " .. buffer(14, 1):uint(), "(" .. opt_no[buffer(14, 1):uint()] .. ")")
	subtree:add(buffer(15, 1), "Reserved: " .. buffer(15, 1):uint())
	subtree:add(buffer(16, 16), "ID: " .. buffer(16, 16))
end

-- Processing by Type (After Base Header)
function dispatch2(code, buffer, tree, type_longname)
	-- Name of break (e.g., Keep Alive)
	local subtree = tree:add(buffer(0), type_longname[code], buffer(0):tvb())

	if code == 2 or code == 8 or code == 21 or code == 23 then
		-- Ack
		-- Keep Alive
		-- Hole Punching
		-- Tunnel Response

		subtree:add(buffer(0, 16), "Hash-based Message Authentication Code: " .. buffer(0, 16))
	elseif code == 3 then
		-- Login Request

		local auth_type = buffer(16, 1):uint()
		local device_id_len = buffer(18, 2):uint()
		local pass_len = buffer(20, 2):uint()
		local cert_len = buffer(22, 2):uint()

		subtree:add(buffer(0, 16), "Application ID: " .. buffer(0, 16))
		subtree:add(buffer(16, 1), "Auth Type: " .. buffer(16, 1):uint())
		subtree:add(buffer(17, 1), "Exc Mode: " .. buffer(17, 1):uint())
		subtree:add(buffer(18, 2), "DeviceID Length: " .. buffer(18, 2):uint())
		subtree:add(buffer(20, 2), "Password Length: " .. buffer(20, 2):uint())
		subtree:add(buffer(22, 2), "Certificate Length: " .. buffer(22, 2):uint())

		if auth_type == 1 then
			-- Basic Authentication

			subtree:add(buffer(24, device_id_len), "DeviceID: " .. buffer(24, device_id_len))
			subtree:add(buffer(24 + device_id_len, pass_len), "Password: " .. buffer(24 + device_id_len, pass_len))
		elseif auth_type == 2 then
			-- Certificate Authentication

			subtree:add(buffer(24, cert_len), "Certificate: " .. buffer(24, cert_len))
		end
	elseif code == 4 then
		-- Login Response

		local comm_key_len = buffer(2, 2):uint()
		local fqdn_len = buffer(8, 2):uint()

		subtree:add(buffer(0, 2), "Cipher Type: " .. buffer(0, 2):uint())
		subtree:add(buffer(2, 2), "Common Key Length: " .. buffer(2, 2):uint())
		subtree:add(buffer(4, 2), "Year: " .. buffer(4, 2):uint())
		subtree:add(buffer(6, 1), "Month: " .. buffer(6, 1):uint())
		subtree:add(buffer(7, 1), "Day: " .. buffer(7, 1):uint())
		subtree:add(buffer(8, 2), "FQDN Length: " .. buffer(8, 2):uint())
		subtree:add(buffer(10, 2), "Padding: " .. buffer(10, 2))
		subtree:add(buffer(12, 4), "Virtual IPv4 Address:" .. buffer(12, 4))
		subtree:add(buffer(16, 16), "Virtual IPv6 Address:" .. buffer(16, 16))
		subtree:add(buffer(32, 4), "NMS's IPv4 Address:" .. buffer(32, 4))
		subtree:add(buffer(36, 16), "NMS's IPv6 Address:" .. buffer(36, 16))
		subtree:add(buffer(52, comm_key_len), "Common Key: " .. buffer(52, comm_key_len))
		subtree:add(buffer(52 + comm_key_len, fqdn_len), "FQDN: " .. buffer(52 + comm_key_len, fqdn_len))
	elseif code == 6 then
		-- Registration Request

		subtree:add(buffer(0, 2), "Application Port: " .. buffer(0, 2))
		subtree:add(buffer(2, 2), "Notification Type: " .. buffer(2, 2):uint())
		subtree:add(buffer(4, 4), "Node Ipv4 Address: " .. buffer(4, 4))
		subtree:add(buffer(8, 16), "Node Ipv6 Address: " .. buffer(8, 16))
		subtree:add(buffer(24, 16), "Hash-based Message Authentication Code: " .. buffer(24, 16))
	elseif code == 7 then
		-- Registration Response

		subtree:add(buffer(0, 1), "NAT Flag: " .. buffer(0, 1):uint():uint())
		subtree:add(buffer(1, 3), "Padding: " .. buffer(1, 3))
		subtree:add(buffer(4, 16), "Hash-based Message Authentication Code: " .. buffer(4, 16))
	elseif code == 11 then
		-- Direction Request

		local fqdn_len = buffer(32, 2):uint()
		local cert_len = buffer(34, 2):uint()

		subtree:add(buffer(0, 16), "Application ID: " .. buffer(0, 16))
		subtree:add(buffer(16, 16), "Path ID: " .. buffer(16, 16))
		subtree:add(buffer(32, 2), "FQDN Length: " .. buffer(32, 2):uint())
		subtree:add(buffer(34, 2), "Certificate Length: " .. buffer(34, 2):uint())
		subtree:add(buffer(36, fqdn_len), "FQDN: " .. buffer(36, fqdn_len))
		subtree:add(buffer(36 + fqdn_len, cert_len), "Certificate: " .. buffer(36 + fqdn_len, cert_len))
		subtree:add(
			buffer(36 + fqdn_len + cert_len, 16),
			"Hash-based Message Authentication Code: " .. buffer(36 + fqdn_len + cert_len, 16)
		)
	elseif code == 12 or code == 13 or code == 14 then
		-- Route Direction To CN
		-- Route Direction Confirmation
		-- Route Direction To MN

		local tunnel_key_len = buffer(116, 2):uint()
		local temp_key_len = buffer(118, 2):uint()
		local fqdn_init_len = buffer(126, 2):uint()
		local fqdn_res_len = buffer(128, 2):uint()

		subtree:add(buffer(0, 16), "Path ID: " .. buffer(0, 16))
		subtree:add(buffer(16, 1), "Process Code: " .. buffer(16, 1):uint())
		subtree:add(buffer(17, 1), "Exc Mode: " .. buffer(17, 1):uint())
		subtree:add(buffer(18, 2), "General Node Flag: " .. buffer(18, 2):uint())
		subtree:add(buffer(20, 2), "NAT initiator Port: " .. buffer(20, 2):uint())
		subtree:add(buffer(22, 2), "NAT responder Port: " .. buffer(22, 2):uint())
		subtree:add(buffer(24, 4), "NAT initiator IPv4: " .. buffer(24, 4))
		subtree:add(buffer(28, 16), "NAT initiator IPv6: " .. buffer(28, 16))
		subtree:add(buffer(44, 4), "NAT responder IPv4: " .. buffer(44, 4))
		subtree:add(buffer(48, 16), "NAT responder IPv6: " .. buffer(48, 16))
		subtree:add(buffer(54, 4), "Initiator Real IPv4: " .. buffer(54, 4))
		subtree:add(buffer(58, 16), "Initiator Real IPv6: " .. buffer(58, 16))
		subtree:add(buffer(74, 4), "Responder Real IPv4: " .. buffer(74, 4))
		subtree:add(buffer(78, 16), "Responder Real IPv6: " .. buffer(78, 16))
		subtree:add(buffer(94, 4), "TRS IPv4: " .. buffer(94, 4))
		subtree:add(buffer(98, 16), "TRS IPv6: " .. buffer(98, 16))
		subtree:add(buffer(114, 2), "Tunnel Key Cipher Type: " .. buffer(114, 2):uint())
		subtree:add(buffer(116, 2), "Tunnel Key Length: " .. buffer(116, 2):uint())
		subtree:add(buffer(118, 2), "Temporary Key Length: " .. buffer(118, 2):uint())
		subtree:add(buffer(120, 2), "Padding: " .. buffer(120, 2))
		subtree:add(buffer(122, 2), "Year: " .. buffer(122, 2):uint())
		subtree:add(buffer(124, 1), "Month: " .. buffer(124, 1):uint())
		subtree:add(buffer(125, 1), "Day: " .. buffer(125, 1):uint())
		subtree:add(buffer(126, 2), "FQDN initiator Length: " .. buffer(126, 2):uint())
		subtree:add(buffer(128, 2), "FQDN responder Length: " .. buffer(128, 2):uint())
		subtree:add(buffer(130, tunnel_key_len), "Tunnel Key: " .. buffer(130, tunnel_key_len))
		subtree:add(
			buffer(130 + tunnel_key_len, temp_key_len),
			"Temporary Key: " .. buffer(130 + tunnel_key_len, temp_key_len)
		)
		subtree:add(
			buffer(130 + tunnel_key_len + temp_key_len, fqdn_init_len),
			"Initiator FQDN: " .. buffer(130 + tunnel_key_len + temp_key_len, fqdn_init_len)
		)
		subtree:add(
			buffer(130 + tunnel_key_len + temp_key_len + fqdn_init_len, fqdn_res_len),
			"Responder FQDN: " .. buffer(130 + tunnel_key_len + temp_key_len + fqdn_init_len, fqdn_res_len)
		)
		subtree:add(
			buffer(130 + tunnel_key_len + temp_key_len + fqdn_init_len + fqdn_res_len, 16),
			"Hash-based Message Authentication Code: "
				.. buffer(130 + tunnel_key_len + temp_key_len + fqdn_init_len + fqdn_res_len, 16)
		)
	elseif code == 17 or code == 18 then
		-- Relay Request
		-- Relay Response

		local tunnel_key_len = buffer(18, 2):uint()

		subtree:add(buffer(0, 16), "Path ID: " .. buffer(0, 16))
		subtree:add(buffer(16, 2), "Tunnel Key Cipher Type: " .. buffer(16, 2):uint())
		subtree:add(buffer(18, 2), "Tunnel Key Length: " .. buffer(18, 2):uint())
		subtree:add(buffer(20, tunnel_key_len), "Tunnel Key: " .. buffer(20, tunnel_key_len))
		subtree:add(
			buffer(20 + tunnel_key_len, 16),
			"Hash-based Message Authentication Code: " .. buffer(20 + tunnel_key_len, 16)
		)
	elseif code == 22 then
		-- Tunnel Request

		local end_key_len = buffer(0, 2):uint()

		subtree:add(buffer(0, 2), "End Key Length: " .. buffer(0, 2):uint())
		subtree:add(buffer(2, 2), "Padding: " .. buffer(2, 2))
		subtree:add(buffer(4, end_key_len), "End Key: " .. buffer(4, end_key_len))
		subtree:add(
			buffer(4 + end_key_len, 16),
			"Hash-based Message Authentication Code: " .. buffer(4 + end_key_len, 16)
		)
	elseif code == 24 then
		-- Capsule Message

		local payload_len = buffer(0, 2):uint()

		subtree:add(buffer(0, 2), "Payload Length: " .. buffer(0, 2):uint())
		subtree:add(buffer(2, 2), "Padding: " .. buffer(2, 2))
		subtree:add(buffer(4, payload_len), "Payload: " .. buffer(4, payload_len))
		subtree:add(
			buffer(4 + payload_len, 16),
			"Hash-based Message Authentication Code: " .. buffer(4 + payload_len, 16)
		)
	end
end
