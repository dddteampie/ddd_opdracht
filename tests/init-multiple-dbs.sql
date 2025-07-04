DO
$$
BEGIN
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ecd') THEN
      CREATE DATABASE ecd;
   END IF;
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'behoeftebepaling') THEN
      CREATE DATABASE behoeftebepaling;
   END IF;
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'aanvraagverwerking') THEN
      CREATE DATABASE aanvraagverwerking;
   END IF;
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'products') THEN
      CREATE DATABASE products;
   END IF;
   IF NOT EXISTS (SELECT FROM pg_database WHERE datname = 'recommendation') THEN
      CREATE DATABASE recommendation;
   END IF;
END
$$;