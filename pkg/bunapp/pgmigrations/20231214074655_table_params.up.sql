update grid_items
set params = jsonb_set(params, '{itemsPerPage}', '5', true)
where type = 'table';

--bun:split

update grid_items
set params = jsonb_set(params, '{denseTable}', 'false', true)
where type = 'table';
