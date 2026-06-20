import { Login, useLogin } from "react-admin";
import { Button, Card, CardContent, Typography } from "@mui/material";
import Logo from "@core/logo/Logo";
import { useAppConfig } from "@core/config/useConfig";

export const OidcLogin = () => {
  const login = useLogin();

  const { auth } = useAppConfig();

  const name = auth?.name || "IDP";

  return (
    <Login>
      <Card
        sx={{
          minWidth: 320,
          textAlign: "center",
          p: 3,
          backgroundColor: "background.paper",
          boxShadow: 4,
        }}
      >
        <CardContent
          sx={{
            display: "flex",
            flexDirection: "column",
            alignItems: "center",
          }}
        >
          <Logo width={64} height={64} />

          <Typography variant="h5" component="h1" gutterBottom>
            Billedapparat
          </Typography>

          <Typography variant="body2" color="text.secondary" sx={{ mb: 4 }}>
            Please login with your Identity Provider to access the admin panel.
          </Typography>

          <Button
            variant="contained"
            color="primary"
            fullWidth
            onClick={() => login({})}
          >
            Login with {name}
          </Button>
        </CardContent>
      </Card>
    </Login>
  );
};
