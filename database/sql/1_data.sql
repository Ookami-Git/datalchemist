INSERT INTO users (name, type, parameters)
VALUES ("admin", "local", '{"password": "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918"}');

INSERT INTO groups (name, description)
VALUES ("admin", "Administrator");

INSERT INTO paramerters (name, value)
VALUES
    ("name","datalchemist"),
    ("lang","en"),
    ("menu","- name: HOME
  link: /
- name: Div example
  link: ---
- name: menu
  subitems:
    - name: view
      link: /view/viewname
    - name: Div example
      link: ---
    - name: view2 with parameters
      link: /view/viewid&value=test
- name: external site
  link: http://www.externalsite.com
  newtab: True"),
    ("theme","light"),
    ("bg_color_light","rgb(142, 114, 173)"),
    ("bg_color2_light","rgb(94, 130, 192)"),
    ("bg_color_dark","rgb(60, 11, 111)"),
    ("bg_color2_dark","rgb(15, 45, 97)"),
    ("ldap", "false"),
    ("ldap_config", "{}");
