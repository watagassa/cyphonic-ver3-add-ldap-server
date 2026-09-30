"use client";

import { ChangeEventHandler, FormEventHandler, useState } from "react";
import axios from "axios";
import {
  Button,
  Checkbox,
  FormControl,
  FormLabel,
  Input,
  NumberDecrementStepper,
  NumberIncrementStepper,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  Text,
  Box,
} from "@chakra-ui/react";

type DeviceData = {
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
};
import { endpoint } from "@/config";

export default function PostDevice() {
  const [formData, setFormData] = useState<DeviceData>({
    version: "1.0",
    device_id: "1f0b1fb5663eb66f0a5155606f91e37f6333263716df86f451a273c0",
    account_id: 1,
    device_name: "test-device",
    fqdn: "1f0b1fb5663eb66f0a5155606f91e37f6333263716df86f451a273c0",
    node_management_area_id: 1,
    device_type: 1,
    adapter_flag: false,
    general_node_flag: false,
    inter_node_certification_flag: false,
    status: 0,
  });
  const handleInput: ChangeEventHandler<
    HTMLInputElement | HTMLTextAreaElement
  > = (event) => {
    const { name, value } = event.currentTarget;
    setFormData({ ...formData, [name]: value });
  };
  const handleSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    axios
      .post(`${endpoint}/api/devices`, formData, {
        headers: { "Content-Type": "application/json" },
      })
      .then((res) => {
        console.log(res);
        window.alert("Success!!");
      })
      .catch((err) => {
        console.log(err);
        window.alert("Failed");
      });
  };

  return (
    <Box maxW="md" mx="auto" mt={8} p={4} borderWidth={1} borderRadius={8}>
      <form onSubmit={handleSubmit}>
        <FormControl>
          <FormLabel htmlFor="version">Version</FormLabel>
          <Input type="text" id="version" name="version" defaultValue={"1.0"} />
        </FormControl>
        <FormControl>
          <FormLabel htmlFor="device_id">Device ID</FormLabel>
          <Input
            type="text"
            id="device_id"
            name="device_id"
            value={formData.device_id}
            onChange={handleInput}
          />
        </FormControl>
        <FormControl>
          <FormLabel margin="0" htmlFor="account_id">
            Account ID
          </FormLabel>
          <Text fontSize="sm" color="red">
            Must be an existing account.
          </Text>
          <NumberInput
            id="account_id"
            name="account_id"
            value={formData.account_id}
            onChange={(value) =>
              setFormData((prevState) => ({
                ...prevState,
                account_id: Number(value),
              }))
            }
            min={1}
          >
            <NumberInputField />
            <NumberInputStepper>
              <NumberIncrementStepper />
              <NumberDecrementStepper />
            </NumberInputStepper>
          </NumberInput>
        </FormControl>
        <FormControl>
          <FormLabel margin="0" htmlFor="device_name">
            Device Name
          </FormLabel>
          <Text fontSize="sm" color="red">
            Must be unique.
          </Text>
          <Input
            type="text"
            id="device_name"
            name="device_name"
            value={formData.device_name}
            onChange={handleInput}
          />
        </FormControl>
        <FormControl>
          <FormLabel margin="0" htmlFor="fqdn">
            FQDN
          </FormLabel>
          <Text fontSize="sm" color="red">
            Must be unique.
          </Text>
          <Input
            type="text"
            id="fqdn"
            name="fqdn"
            value={formData.fqdn}
            onChange={handleInput}
          />
        </FormControl>
        <FormControl>
          <FormLabel htmlFor="node_management_area_id">
            Node Management Area ID
          </FormLabel>
          <NumberInput
            id="node_management_area_id"
            name="node_management_area_id"
            value={formData.node_management_area_id}
            onChange={(value) =>
              setFormData((prevState) => ({
                ...prevState,
                node_management_area_id: Number(value),
              }))
            }
            min={1}
          >
            <NumberInputField />
            <NumberInputStepper>
              <NumberIncrementStepper />
              <NumberDecrementStepper />
            </NumberInputStepper>
          </NumberInput>
        </FormControl>
        <FormControl>
          <FormLabel margin="0" htmlFor="device_type">
            Device Type
          </FormLabel>
          <Text fontSize="sm" color="gray">
            1: Linux / 2: Windows / 3: Darwin / 4: iOS / 5: Android
          </Text>
          <NumberInput
            id="device_type"
            name="device_type"
            value={formData.device_type}
            onChange={(value) =>
              setFormData((prevState) => ({
                ...prevState,
                device_type: Number(value),
              }))
            }
            min={1}
            max={5} // https://github.com/Pluslab/cyphonic/blob/develop/.docker/db/init.d/001-administrative-dml.sql#L6-L10
          >
            <NumberInputField />
            <NumberInputStepper>
              <NumberIncrementStepper />
              <NumberDecrementStepper />
            </NumberInputStepper>
          </NumberInput>
        </FormControl>
        <FormControl>
          <FormLabel htmlFor="adapter_flag">Adapter Flag</FormLabel>
          <Checkbox
            id="adapter_flag"
            name="adapter_flag"
            isChecked={formData.adapter_flag}
            onChange={(event) =>
              setFormData((prevState) => ({
                ...prevState,
                adapter_flag: event.target.checked,
              }))
            }
          />
        </FormControl>

        <FormControl>
          <FormLabel htmlFor="general_node_flag">General Node Flag</FormLabel>
          <Checkbox
            id="general_node_flag"
            name="general_node_flag"
            isChecked={formData.general_node_flag}
            onChange={(event) =>
              setFormData((prevState) => ({
                ...prevState,
                general_node_flag: event.target.checked,
              }))
            }
          />
        </FormControl>

        <FormControl>
          <FormLabel htmlFor="inter_node_certification_flag">
            Inter Node Certification Flag
          </FormLabel>
          <Checkbox
            id="inter_node_certification_flag"
            name="inter_node_certification_flag"
            isChecked={formData.inter_node_certification_flag}
            onChange={(event) =>
              setFormData((prevState) => ({
                ...prevState,
                inter_node_certification_flag: event.target.checked,
              }))
            }
          />
        </FormControl>
        <FormControl>
          <FormLabel htmlFor="status">Status</FormLabel>
          <NumberInput
            id="status"
            name="status"
            value={formData.status}
            onChange={(value) =>
              setFormData((prevState) => ({
                ...prevState,
                status: Number(value),
              }))
            }
            min={0}
            max={3} // TODO: account statusはチェックボックス方式にするかも！(このフィールドの仕様は検討中)
          >
            <NumberInputField />
            <NumberInputStepper>
              <NumberIncrementStepper />
              <NumberDecrementStepper />
            </NumberInputStepper>
          </NumberInput>
        </FormControl>
        <Button type="submit" mt={4}>
          Submit
        </Button>
      </form>
    </Box>
  );
}
