// src/routes.js (Aangepast)

const routes = [
  {
    path: "/",
    exact: true,
    component: "WelcomePage",
  },
  {
    path: "/clients",
    exact: true,
    component: "ClientSelectionPage",
  },
  {
    path: "/clients/:clientId",
    exact: true,
    component: "ClientDetailPage", 
  },
  {
    path: "/clients/:clientId/behoeften",
    exact: false, 
    component: "BehoeftebepalingPage",
  },
  {
    path: "/clients/:clientId/aanvraag/:aanvraagId/advies",
    exact: true,
    component: "AdviesPage",
  },
  {
    path: "/products",
    exact: true,
    component: "ProductListPage",
  },
  {
    path: "/products/:ean",
    exact: true,
    component: "ProductDetailPage",
  },
];

export default routes;