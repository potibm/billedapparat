import React, { ReactNode } from "react";
import { Card } from "flowbite-react";
import Logo from "@core/logo/Logo";
import { useAppConfig } from "@core/config/useConfig";

interface BaseCardProps {
  children: ReactNode;
  title?: string;
}

const BaseCard: React.FC<BaseCardProps> = ({ children, title = null }) => {
  const { version } = useAppConfig();

  return (
    <div className="flex justify-center items-center h-screen">
      <Card className="max-w-sm bg-white/15 border-color-[#00F5FF]/90 backdrop-blur-sm">
        <h5 className="flex items-center gap-2 text-2xl font-bold tracking-tight text-white">
          <Logo className="h-7 w-7 text-[#00F5FF]" />
          Billedapparat
        </h5>
        <div className="my-3 text-white">
          {title && <h2 className="text-xl mb-2">{title}</h2>}

          <div className="text-gray-300">{children}</div>
        </div>

        <div className="my-3 smaller border-t-gray-600 border-t">
          Version:{" "}
          <a
            href="https://github.com/potibm/billedapparat/releases"
            target="_blank"
            rel="noopener noreferrer"
          >
            {version}
          </a>
        </div>
      </Card>
    </div>
  );
};

export default BaseCard;
