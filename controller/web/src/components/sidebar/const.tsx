import {
  HiOutlineDeviceMobile,
  HiOutlineDeviceTablet,
  HiOutlineUser,
} from "react-icons/hi";

import { LinkItemProps } from "@/components/sidebar/types";

export const LinkItems: Array<LinkItemProps> = [
  { name: "HOME", icon: null, href: "/" },
  { name: "GET", icon: null, href: "/" },
  { name: "DEVICE", icon: HiOutlineDeviceMobile, href: "/get/device" },
  { name: "DEVICES", icon: HiOutlineDeviceMobile, href: "/get/devices" },
  {
    name: "GENERAL DEVICE",
    icon: HiOutlineDeviceTablet,
    href: "/get/general_device",
  },
  { name: "ACCOUNT", icon: HiOutlineUser, href: "/get/account" },
  { name: "POST", icon: null, href: "/" },
  { name: "DEVICE", icon: HiOutlineDeviceMobile, href: "/post/device" },
];
