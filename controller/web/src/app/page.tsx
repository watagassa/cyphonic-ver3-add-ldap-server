"use client";

import {
  Box,
  Container,
  Flex,
  Heading,
  Icon,
  Stack,
  Text,
  useColorModeValue,
} from "@chakra-ui/react";
import { ReactElement } from "react";
import {
  FcBusinessman,
  FcCollaboration,
  FcTabletAndroid,
  FcUpload,
} from "react-icons/fc";

interface CardProps {
  heading: string;
  description: string;
  icon: ReactElement;
  href: string;
}
const cards = [
  {
    heading: "DEVICE",
    icon: <Icon as={FcTabletAndroid} w={10} h={10} />,
    description: "",
    href: "/get/device",
  },
  {
    heading: "DEVICES",
    icon: <Icon as={FcTabletAndroid} w={10} h={10} />,
    description: "",
    href: "/get/devices",
  },
  {
    heading: "GENERAL DEVICE",
    icon: <Icon as={FcCollaboration} w={10} h={10} />,
    description: "",
    href: "/get/general_device",
  },
  {
    heading: "ACCOUNT",
    icon: <Icon as={FcBusinessman} w={10} h={10} />,
    description: "",
    href: "/get/account",
  },
  {
    heading: "POST DEVICE",
    icon: <Icon as={FcUpload} w={10} h={10} />,
    description: "",
    href: "/post/device",
  },
];

const Card = ({ heading, description, icon, href = "" }: CardProps) => {
  return (
    <Box
      as="div"
      maxW={{ base: "full", md: "275px" }}
      w={"full"}
      borderWidth="1px"
      borderRadius="lg"
      overflow="hidden"
      p={5}
      transition="transform 0.2s"
      _hover={{
        transform: "translateY(-4px)",
      }}
      onClick={() => {
        if (href) {
          window.location.href = href;
        }
      }}
    >
      <Stack align={"center"} spacing={2}>
        <Flex
          w={16}
          h={16}
          align={"center"}
          justify={"center"}
          color={"white"}
          rounded={"full"}
          bg={useColorModeValue("gray.100", "gray.700")}
        >
          {icon}
        </Flex>
        <Box mt={2}>
          <Heading size="sm">{heading}</Heading>
          <Text mt={1} fontSize={"sm"} noOfLines={2}>
            {description}
          </Text>
        </Box>
        {href && (
          <Box
            mt={4}
            w={"full"}
            bg={"blue.600"}
            color={"white"}
            rounded={"md"}
            px={6}
            py={2}
            fontWeight={"bold"}
            _hover={{
              bg: "blue.700",
            }}
            onClick={() => {
              window.location.href = href;
            }}
            as="button"
          >
            Open
          </Box>
        )}
      </Stack>
    </Box>
  );
};

export default function Home() {
  return (
    <Box p={4}>
      <Stack spacing={4} as={Container} maxW={"3xl"} textAlign={"center"}>
        <Heading fontSize={{ base: "2xl", sm: "4xl" }} fontWeight={"bold"}>
          CYPHONIC WEB CONTROLLER
        </Heading>
        <Text color={"gray.600"} fontSize={{ base: "sm", sm: "lg" }}>
          クラウドを操作するためのソフトウェアコントローラ
        </Text>
      </Stack>

      <Container maxW={"5xl"} mt={12}>
        <Flex flexWrap="wrap" gridGap={6} justify="center">
          {cards.map((card, index) => (
            <Card
              key={index}
              heading={card.heading}
              icon={card.icon}
              description={card.description}
              href={card.href}
            />
          ))}
        </Flex>
      </Container>
    </Box>
  );
}
