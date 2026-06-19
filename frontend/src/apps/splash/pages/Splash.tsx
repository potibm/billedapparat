import React from "react";
import BaseCard from "../components/BaseCard";
import AdminPanelSettingsIcon from "@mui/icons-material/AdminPanelSettings";
import ViewCarouselIcon from "@mui/icons-material/ViewCarousel";
import { Link } from "react-router-dom";

export const Splash: React.FC = () => {
  return (
    <BaseCard>
      <>
        <p className="mb-2">A screen director for demoparties</p>

        <Link to="/beamer">
          <ViewCarouselIcon fontSize="large" color="inherit" className="m-2" />
          Beamer
        </Link>

        <br />

        <Link to="/admin">
          <AdminPanelSettingsIcon
            fontSize="large"
            color="inherit"
            className="m-2"
          />
          Admin
        </Link>
      </>
    </BaseCard>
  );
};

export default Splash;
