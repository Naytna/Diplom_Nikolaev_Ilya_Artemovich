update learning.users as u
set password_hash = '$2a$10$Am7pYLO.rTA7JR/R6v2ek.0IHku6orwM6rEGsOEGBoyvehwp5Us5i'
from learning.roles as r
where r.id = u.role_id
  and r.code = 'expert';

update learning.users as u
set password_hash = '$2a$10$FxM3LCuF.Y6m3niPzCQNquiUv3uLNUzUGSAhO7opFVdQh3g7X8Jne'
from learning.roles as r
where r.id = u.role_id
  and r.code = 'student';
