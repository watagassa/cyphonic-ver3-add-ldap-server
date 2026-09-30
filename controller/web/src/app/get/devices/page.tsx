"use client";

import axios, { AxiosResponse } from "axios";
import useSWR from "swr";
import { Box, FormControl, FormLabel, Input, Text } from "@chakra-ui/react";
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

export default function GetAllDevice() {
  const { data, error } = useSWR<AxiosResponse<DeviceData[]>, Error>(
    `${endpoint}/api/devices`,
    axios.get
  );

  if (error) return <div>failed to load</div>;

  return (
    <>
      <Box>
        {data && data.data && data.data.length > 0 ? (
          data.data.map((device) => (
            <div key={device.id}>
              <Text>{device.adapter_flag}</Text>
              <Text>ID: {device.id}</Text>
              <Text>Version: {device.version}</Text>
              <Text>Device ID: {device.device_id}</Text>
              <Text>Account ID: {device.account_id}</Text>
              <Text>Device Name: {device.device_name}</Text>
              <Text>FQDN: {device.fqdn}</Text>
              <Text>
                Node Management Area ID: {device.node_management_area_id}
              </Text>
              <Text>Device Type: {device.device_type}</Text>
              <Text>Adapter Flag: {device.adapter_flag.toString()}</Text>
              <Text>
                General Node Flag: {device.general_node_flag.toString()}
              </Text>
              <Text>
                Inter Node Certification Flag:{" "}
                {device.inter_node_certification_flag.toString()}
              </Text>
              <Text>Status: {device.status}</Text>
              <Text>Created At: {device.created_at}</Text>
              <Text>Updated At: {device.updated_at}</Text>
              <br />
            </div>
          ))
        ) : (
          <Text>Not Found</Text>
        )}
      </Box>
    </>
  );
}
