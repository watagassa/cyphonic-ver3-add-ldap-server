"use client";

import axios, { AxiosResponse } from "axios";
import useSWR from "swr";
import { Box, FormControl, FormLabel, Input, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { endpoint } from "@/config";

type DeviceData = {
  id: number;
  version: string;
  device_id: string;
  account_id: number;
  device_name: string;
  fqdn: string;
  node_management_area_id: number;
  device_type: number;
  adapter_flag: boolean;
  general_node_flag: boolean;
  inter_node_certification_flag: boolean;
  status: number;
  created_at: string;
  updated_at: string;
  message?: string;
};

function useDebounce(value: string | null, delay: number) {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);

    return () => {
      clearTimeout(timer);
    };
  }, [value, delay]);

  return debouncedValue;
}

export default function GetDevice() {
  const [deviceId, setDeviceId] = useState("");
  const debouncedDeviceId = useDebounce(deviceId, 500);

  const { data, error } = useSWR<AxiosResponse<DeviceData>, Error>(
    debouncedDeviceId ? `${endpoint}/api/device?device_id=${deviceId}` : null,
    axios.get
  );

  if (error) return <div>failed to load</div>;

  return (
    <>
      <FormControl>
        <FormLabel>Device ID</FormLabel>
        <Input
          type="text"
          value={deviceId}
          onChange={(e) => setDeviceId(e.target.value)}
        />
      </FormControl>

      <Box>
        <Text>
          message:{" "}
          {data?.data?.message
            ? data.data.message
            : data?.data
            ? "found"
            : "no matching records found"}
        </Text>
        <Text>ID: {data?.data.id}</Text>
        <Text>Version: {data?.data.version}</Text>
        <Text>Device ID: {data?.data.device_id}</Text>
        <Text>Account ID: {data?.data.account_id}</Text>
        <Text>Device Name: {data?.data.device_name}</Text>
        <Text>FQDN: {data?.data.fqdn}</Text>
        <Text>
          Node Management Area ID: {data?.data.node_management_area_id}
        </Text>
        <Text>Device Type: {data?.data.device_type}</Text>
        <Text>Adapter Flag: {data?.data.adapter_flag?.toString()}</Text>
        <Text>
          General Node Flag: {data?.data.general_node_flag?.toString()}
        </Text>
        <Text>
          Inter Node Certification Flag:{" "}
          {data?.data.inter_node_certification_flag?.toString()}
        </Text>
        <Text>Status: {data?.data.status}</Text>
        <Text>Created At: {data?.data.created_at}</Text>
        <Text>Updated At: {data?.data.updated_at}</Text>
      </Box>
    </>
  );
}
