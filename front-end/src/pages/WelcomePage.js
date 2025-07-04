import React from 'react';

const WelcomePage = () => {
  return (
    <div style={{
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'center',
      alignItems: 'center',
      minHeight: '100vh',
      textAlign: 'center',
      padding: '20px'
    }}>
      <h1>Welkom bij jouw Landy SPA!</h1>
      <p>Dit is een simpele pagina om te laten zien dat je Landy template werkt.</p>
      <p>Je kunt hier beginnen met het bouwen van je geweldige frontend.</p>
    </div>
  );
};

export default WelcomePage;