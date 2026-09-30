"use client";

import axios, { AxiosResponse } from "axios";
import useSWR from "swr";
import { Box, FormControl, FormLabel, Input, Text } from "@chakra-ui/react";
import { useEffect, useState } from "react";
import { endpoint } from "@/config";

type GeneralDevice = {
  id: number;
  adapter_id: string;
  device_id: string;
  mac_address: string;
  enable_ipv6: boolean;
  auth_type: number;
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

export default function GetGeneralDevice() {
  const [deviceId, setDeviceId] = useState("");
  const debouncedDeviceId = useDebounce(deviceId, 500);

  const { data, error } = useSWR<AxiosResponse<GeneralDevice>, Error>(
    debouncedDeviceId
      ? `${endpoint}/api/general_device?device_id=${debouncedDeviceId}`
      : null,
    axios.get
  );

  console.log(data);

  if (error) return <div>failed to load</div>;

  return (
    <>
      <FormControl>
        <FormLabel>General Device ID</FormLabel>
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
        <Text>Adapter ID: {data?.data.adapter_id}</Text>
        <Text>Device ID: {data?.data.device_id}</Text>
        <Text>MAC Address: {data?.data.mac_address}</Text>
        <Text>Enable IPv6: {data?.data.enable_ipv6?.toString()}</Text>
        <Text>Auth Type: {data?.data.auth_type}</Text>
        <Text>Created At: {data?.data.created_at}</Text>
        <Text>Updated At: {data?.data.updated_at}</Text>
      </Box>
    </>
  );
}
