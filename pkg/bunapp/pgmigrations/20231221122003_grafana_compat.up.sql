alter table projects
add column prom_compat boolean not null default false;

--bun:split

update grid_items
set width = 2 * width, x_axis = 2 * x_axis;
