import React, { ReactNode } from "react";
import { Card } from "flowbite-react";

interface BaseCardProps {
  children: ReactNode;
  title?: string;
}

const BaseCard: React.FC<BaseCardProps> = ({ children, title = null }) => {
  return (
    <div className="flex justify-center items-center h-screen">
      <Card className="max-w-sm bg-white/15 border-color-[#00F5FF]/90 backdrop-blur-sm">
        <h5 className="text-2xl font-bold tracking-tight text-[#00F5FF]">
          <img
            src="/web-app-manifest-192x192.png"
            alt="Billedapparat"
            className="align-text-top h-7 inline"
          />{" "}
          Billedapparat
        </h5>

        <div className="my-3 text-white">
          {title && <h2 className="text-xl mb-2">{title}</h2>}

          {children}
        </div>
      </Card>
    </div>
  );
};

export default BaseCard;
