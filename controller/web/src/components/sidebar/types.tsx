import { BoxProps, FlexProps } from "@chakra-ui/react";
import { ReactNode } from "react";
import { IconType } from "react-icons";

export interface SidebarProps extends BoxProps {
  onClose: () => void;
}

export interface NavItemProps extends FlexProps {
  icon: IconType | null;
  children: ReactNode;
  href: string;
}

export interface MobileProps extends FlexProps {
  onOpen: () => void;
}

export interface LinkItemProps {
  name: string;
  icon: IconType | null;
  href: string;
}
