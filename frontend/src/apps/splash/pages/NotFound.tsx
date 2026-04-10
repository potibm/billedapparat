import React from "react";
import BaseCard from "../components/BaseCard";

export const NotFound: React.FC = () => {
  return (
    <BaseCard title="404">
      <p>Page not found.</p>
    </BaseCard>
  );
};

export default NotFound;
