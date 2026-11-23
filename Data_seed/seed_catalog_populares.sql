-- Carga de dados (seed) do catálogo com autores e obras atuais e populares.
-- Versão para MariaDB (10.2.1+ ou superior, onde as CHECK constraints são aplicadas).
-- Tabelas: authors, publishers, genres, works, work_authors, work_genres, editions.
-- Os IDs (PK) são gerados pelo AUTO_INCREMENT. As tabelas de relação e as edições
-- encontram os IDs pelo nome/título, então o script funciona mesmo que as tabelas
-- já tenham dados. Rodar duas vezes duplica autores, editoras, obras e edições
-- (genres tem nome UNIQUE e falharia).
-- ATENÇÃO: os ISBNs são fictícios (o schema só valida os 13 caracteres) e
-- páginas/datas das edições brasileiras são aproximadas. Ajuste se precisar de dados exatos.
-- Uso: mariadb -u USUARIO -p NOME_DO_BANCO < seed_catalog_populares_mariadb.sql

-- Garante a codificação correta dos acentos, independente do padrão do cliente.
SET NAMES utf8mb4 COLLATE utf8mb4_unicode_ci;

START TRANSACTION;

-- ---------------------------------------------------------------- authors
-- birth_date NULL quando a data exata não é conhecida/publicada.
INSERT INTO authors (name, biography, birth_date, death_date, country_of_origin, sex) VALUES
('J.K. Rowling', 'Autora britânica da série Harry Potter, uma das franquias literárias mais vendidas da história.', '1965-07-31', NULL, 'Reino Unido', 'feminino'),
('Stephen King', 'Autor norte-americano conhecido como mestre do terror e do suspense, com dezenas de adaptações para cinema e TV.', '1947-09-21', NULL, 'Estados Unidos', 'masculino'),
('Colleen Hoover', 'Autora norte-americana de romances contemporâneos, fenômeno de vendas e do BookTok.', '1979-12-11', NULL, 'Estados Unidos', 'feminino'),
('Brandon Sanderson', 'Autor norte-americano de fantasia épica, criador do Cosmere e da série O Arquivo das Tempestades.', '1975-12-19', NULL, 'Estados Unidos', 'masculino'),
('Rick Riordan', 'Autor norte-americano de fantasia juvenil inspirada em mitologias, criador de Percy Jackson.', '1964-06-05', NULL, 'Estados Unidos', 'masculino'),
('Taylor Jenkins Reid', 'Autora norte-americana de romances de ficção histórica e contemporânea, best-seller internacional.', '1983-12-20', NULL, 'Estados Unidos', 'feminino'),
('Itamar Vieira Junior', 'Escritor e geógrafo baiano, autor de Torto Arado, vencedor do Prêmio Jabuti e do Leya.', NULL, NULL, 'Brasil', 'masculino'),
('Freida McFadden', 'Autora norte-americana de thrillers psicológicos e médica, best-seller com a série A Empregada.', NULL, NULL, 'Estados Unidos', 'feminino'),
('Sarah J. Maas', 'Autora norte-americana de fantasia romântica, criadora das séries Corte de Espinhos e Rosas e Trono de Vidro.', '1986-03-05', NULL, 'Estados Unidos', 'feminino'),
('Paulo Coelho', 'Escritor brasileiro, autor de O Alquimista, um dos livros mais traduzidos do mundo.', '1947-08-24', NULL, 'Brasil', 'masculino'),
('Andy Weir', 'Autor norte-americano de ficção científica hard, conhecido por O Marciano e Projeto Hail Mary.', '1972-06-16', NULL, 'Estados Unidos', 'masculino');

-- -------------------------------------------------------------- publishers
INSERT INTO publishers (name, website_url, country_of_origin) VALUES
('Rocco',         'https://www.rocco.com.br', 'Brasil'),
('Suma',          'https://www.companhiadasletras.com.br', 'Brasil'),
('Galera Record', 'https://www.record.com.br', 'Brasil'),
('Arqueiro',      'https://www.editoraarqueiro.com.br', 'Brasil'),
('Intrínseca',    'https://www.intrinseca.com.br', 'Brasil'),
('Paralela',      'https://www.companhiadasletras.com.br', 'Brasil'),
('Todavia',       'https://www.todavialivros.com.br', 'Brasil');

-- ------------------------------------------------------------------ genres
INSERT INTO genres (name, active) VALUES
('Fantasia', TRUE),
('Romance', TRUE),
('Terror', TRUE),
('Suspense', TRUE),
('Ficção Científica', TRUE),
('Young Adult', TRUE),
('Ficção Contemporânea', TRUE),
('Romance Literário', TRUE),
('Distopia', FALSE);

-- ------------------------------------------------------------------- works
INSERT INTO works (title, synopsis, original_language, first_publication_date, status, minimum_age) VALUES
('Harry Potter e a Pedra Filosofal', 'Um garoto órfão descobre ser bruxo e começa a estudar na Escola de Magia e Bruxaria de Hogwarts.', 'en', '1997-06-26', 'published', 10),
('It: A Coisa', 'Um grupo de amigos enfrenta, na infância e na vida adulta, uma entidade que assume a forma de seus piores medos em Derry.', 'en', '1986-09-15', 'published', 16),
('É Assim que Acaba', 'Lily se apaixona por um neurocirurgião e precisa enfrentar a complexidade de um relacionamento abusivo.', 'en', '2016-08-02', 'published', 16),
('O Caminho dos Reis', 'Primeiro volume de O Arquivo das Tempestades, em um mundo assolado por tempestades e guerras por lâminas lendárias.', 'en', '2010-08-31', 'published', 14),
('Percy Jackson e o Ladrão de Raios', 'Percy descobre ser filho de um deus grego e é acusado de roubar o raio de Zeus.', 'en', '2005-06-28', 'published', 10),
('Os Sete Maridos de Evelyn Hugo', 'Uma lendária atriz de Hollywood decide contar a uma jovem jornalista os segredos de sua vida e de seus casamentos.', 'en', '2017-06-13', 'published', 14),
('Torto Arado', 'Duas irmãs, filhas de trabalhadores rurais na Chapada Diamantina, narram a luta pela terra e a herança da escravidão.', 'pt-BR', '2019-05-01', 'published', 16),
('A Empregada', 'Uma jovem com passado misterioso aceita trabalhar como empregada doméstica em uma casa cheia de segredos.', 'en', '2022-04-26', 'published', 16),
('Corte de Espinhos e Rosas', 'Feyre, uma caçadora humana, é levada ao reino das fadas e se envolve em uma ameaça que pode destruir ambos os mundos.', 'en', '2015-05-05', 'published', 18),
('O Alquimista', 'O pastor Santiago parte em busca de um tesouro e descobre o sentido de seguir a própria lenda pessoal.', 'pt-BR', '1988-01-01', 'published', 0),
('Projeto Hail Mary', 'Um professor acorda sozinho em uma nave, sem memória, com a missão de salvar a Terra.', 'en', '2021-05-04', 'published', 12);

-- ------------------------------------------------------------ work_authors
-- Relaciona obra e autor pelo título e nome (os IDs vêm do AUTO_INCREMENT).
INSERT INTO work_authors (work_id, author_id)
SELECT w.id, a.id
FROM (
  SELECT 'Harry Potter e a Pedra Filosofal' AS work_title, 'J.K. Rowling' AS author_name
  UNION ALL SELECT 'It: A Coisa', 'Stephen King'
  UNION ALL SELECT 'É Assim que Acaba', 'Colleen Hoover'
  UNION ALL SELECT 'O Caminho dos Reis', 'Brandon Sanderson'
  UNION ALL SELECT 'Percy Jackson e o Ladrão de Raios', 'Rick Riordan'
  UNION ALL SELECT 'Os Sete Maridos de Evelyn Hugo', 'Taylor Jenkins Reid'
  UNION ALL SELECT 'Torto Arado', 'Itamar Vieira Junior'
  UNION ALL SELECT 'A Empregada', 'Freida McFadden'
  UNION ALL SELECT 'Corte de Espinhos e Rosas', 'Sarah J. Maas'
  UNION ALL SELECT 'O Alquimista', 'Paulo Coelho'
  UNION ALL SELECT 'Projeto Hail Mary', 'Andy Weir'
) m
JOIN works   w ON w.title = m.work_title
JOIN authors a ON a.name  = m.author_name;

-- ------------------------------------------------------------- work_genres
INSERT INTO work_genres (work_id, genre_id)
SELECT w.id, g.id
FROM (
  SELECT 'Harry Potter e a Pedra Filosofal' AS work_title, 'Fantasia' AS genre_name
  UNION ALL SELECT 'Harry Potter e a Pedra Filosofal', 'Young Adult'
  UNION ALL SELECT 'It: A Coisa', 'Terror'
  UNION ALL SELECT 'It: A Coisa', 'Suspense'
  UNION ALL SELECT 'É Assim que Acaba', 'Romance'
  UNION ALL SELECT 'É Assim que Acaba', 'Ficção Contemporânea'
  UNION ALL SELECT 'O Caminho dos Reis', 'Fantasia'
  UNION ALL SELECT 'Percy Jackson e o Ladrão de Raios', 'Fantasia'
  UNION ALL SELECT 'Percy Jackson e o Ladrão de Raios', 'Young Adult'
  UNION ALL SELECT 'Os Sete Maridos de Evelyn Hugo', 'Romance'
  UNION ALL SELECT 'Os Sete Maridos de Evelyn Hugo', 'Ficção Contemporânea'
  UNION ALL SELECT 'Torto Arado', 'Romance Literário'
  UNION ALL SELECT 'A Empregada', 'Suspense'
  UNION ALL SELECT 'Corte de Espinhos e Rosas', 'Fantasia'
  UNION ALL SELECT 'Corte de Espinhos e Rosas', 'Romance'
  UNION ALL SELECT 'O Alquimista', 'Ficção Contemporânea'
  UNION ALL SELECT 'O Alquimista', 'Romance Literário'
  UNION ALL SELECT 'Projeto Hail Mary', 'Ficção Científica'
) m
JOIN works  w ON w.title = m.work_title
JOIN genres g ON g.name  = m.genre_name;

-- ----------------------------------------------------------------- editions
-- Varia os status (archived, published, pending_review, draft). A última edição
-- não tem ISBN e usa isbn_absence_reason (exigido por ck_edition_isbn).
-- A obra e a editora são encontradas pelo título e pelo nome.
INSERT INTO editions (work_id, publisher_id, title, isbn, isbn_absence_reason, edition_number, page_count, publication_date, language, image_url, status, minimum_age)
SELECT w.id, p.id, m.title, m.isbn, m.isbn_absence_reason, m.edition_number, m.page_count, m.publication_date, m.language, m.image_url, m.status, m.minimum_age
FROM (
  SELECT 'Harry Potter e a Pedra Filosofal' AS work_title, 'Rocco' AS publisher_name, 'Harry Potter e a Pedra Filosofal' AS title, '9788500000101' AS isbn, '' AS isbn_absence_reason, 1 AS edition_number, 264 AS page_count, '2000-01-01' AS publication_date, 'pt-BR' AS language, '' AS image_url, 'archived' AS status, 10 AS minimum_age
  UNION ALL SELECT 'Harry Potter e a Pedra Filosofal', 'Rocco', 'Harry Potter e a Pedra Filosofal', '9788500000118', '', 2, 264, '2017-11-27', 'pt-BR', 'https://example.com/capas/hp-pedra-filosofal.jpg', 'published', 10
  UNION ALL SELECT 'It: A Coisa', 'Suma', 'It: A Coisa', '9788500000125', '', 1, 1104, '2014-09-01', 'pt-BR', 'https://example.com/capas/it-a-coisa.jpg', 'published', 16
  UNION ALL SELECT 'É Assim que Acaba', 'Galera Record', 'É Assim que Acaba', '9788500000132', '', 1, 368, '2017-01-01', 'pt-BR', 'https://example.com/capas/e-assim-que-acaba.jpg', 'published', 16
  UNION ALL SELECT 'O Caminho dos Reis', 'Arqueiro', 'O Caminho dos Reis', '9788500000149', '', 1, 1008, '2010-01-01', 'pt-BR', '', 'published', 14
  UNION ALL SELECT 'Percy Jackson e o Ladrão de Raios', 'Intrínseca', 'Percy Jackson e o Ladrão de Raios', '9788500000156', '', 1, 400, '2008-01-01', 'pt-BR', '', 'published', 10
  UNION ALL SELECT 'Os Sete Maridos de Evelyn Hugo', 'Paralela', 'Os Sete Maridos de Evelyn Hugo', '9788500000163', '', 1, 360, '2019-01-01', 'pt-BR', 'https://example.com/capas/evelyn-hugo.jpg', 'published', 14
  UNION ALL SELECT 'Torto Arado', 'Todavia', 'Torto Arado', '9788500000170', '', 1, 264, '2019-05-01', 'pt-BR', 'https://example.com/capas/torto-arado.jpg', 'published', 16
  UNION ALL SELECT 'A Empregada', 'Arqueiro', 'A Empregada', '9788500000187', '', 1, 336, '2023-01-01', 'pt-BR', '', 'published', 16
  UNION ALL SELECT 'Corte de Espinhos e Rosas', 'Galera Record', 'Corte de Espinhos e Rosas', '9788500000194', '', 1, 448, '2016-01-01', 'pt-BR', '', 'published', 18
  UNION ALL SELECT 'O Alquimista', 'Rocco', 'O Alquimista', '9788500000200', '', 1, 208, '1988-01-01', 'pt-BR', '', 'published', 0
  UNION ALL SELECT 'Projeto Hail Mary', 'Intrínseca', 'Projeto Hail Mary', '9788500000217', '', 1, 528, '2021-11-01', 'pt-BR', '', 'pending_review', 12
  UNION ALL SELECT 'Projeto Hail Mary', 'Intrínseca', 'Projeto Hail Mary (pré-lançamento)', NULL, 'Edição de pré-lançamento ainda sem ISBN atribuído', 1, 528, NULL, 'pt-BR', '', 'draft', 12
) m
JOIN works      w ON w.title = m.work_title
JOIN publishers p ON p.name  = m.publisher_name;

COMMIT;