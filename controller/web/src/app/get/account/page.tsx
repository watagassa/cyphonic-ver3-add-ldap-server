"use client";

import axios, { AxiosResponse } from "axios";
import useSWR from "swr";
import {
  Box,
  FormControl,
  FormLabel,
  NumberDecrementStepper,
  NumberIncrementStepper,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  Text,
} from "@chakra-ui/react";
import { useState } from "react";
import { endpoint } from "@/config";

type AccountData = {
  id: number;
  email: string;
  password: string;
  state: number;
  created_at: string;
  updated_at: string;
  message?: string;
};

export default function GetAccount() {
  const [accountId, setAccountId] = useState<string>("0");

  const { data, error } = useSWR<AxiosResponse<AccountData>, Error>(
    `${endpoint}/api/accounts?account_id=${accountId}`,
    axios.get,
    { revalidateOnMount: false }
  );

  if (error) return <div>failed to load</div>;

  return (
    <>
      <FormControl>
        <FormLabel>Account ID</FormLabel>
        <NumberInput min={1} defaultValue={1} onChange={setAccountId}>
          <NumberInputField />
          <NumberInputStepper>
            <NumberIncrementStepper />
            <NumberDecrementStepper />
          </NumberInputStepper>
        </NumberInput>
      </FormControl>

      {data?.data && (
        <Box>
          <Text>ID: {data?.data.id}</Text>
          <Text>Email: {data?.data.email}</Text>
          <Text>Password: {data?.data.password}</Text>
          <Text>State: {data?.data.state}</Text>
          <Text>Created At: {data?.data.created_at}</Text>
          <Text>Updated At: {data?.data.updated_at}</Text>
        </Box>
      )}
    </>
  );
}
