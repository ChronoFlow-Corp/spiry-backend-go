package pgx

const chatWithMediasQuery = `with command_medias_agg as (
    select
        cmm.command_id,
        coalesce(
                        jsonb_agg(
                        jsonb_strip_nulls(
                                jsonb_build_object(
                                        'id', cmm.id,
                                        'name', cmm.name,
                                        'type', cmm.type,
                                        'url', cmm.url,
                                        'size', cmm.size,
                                        'command_id', cmm.command_id,
                                        'user_id', cmm.user_id,
                                        'created_at', cmm.created_at,
                                        'updated_at', cmm.updated_at
                                )
                        )
                                 ) filter (where cmm.id is not null),
                        '[]'::jsonb
        ) as command_medias
    from command_medias cmm
    group by cmm.command_id
),
     result_medias_agg as (
         select
             rm.result_id,
             coalesce(
                             jsonb_agg(
                             jsonb_strip_nulls(
                                     jsonb_build_object(
                                             'id', rm.id,
                                             'name', rm.name,
                                             'type', rm.type,
                                             'url', rm.url,
                                             'size', rm.size,
                                             'result_id', rm.result_id,
                                             'user_id', rm.user_id,
                                             'created_at', rm.created_at,
                                             'updated_at', rm.updated_at
                                     )
                             )
                                      ) filter (where rm.id is not null),
                             '[]'::jsonb
             ) as result_medias
         from result_medias rm
         group by rm.result_id
     ),
     result_per_command as (
         select
             r.command_id,
             jsonb_strip_nulls(
                     jsonb_build_object(
                             'id', r.id,
                             'text', r.text,
                             'command_id', r.command_id,
                             'tool_id', r.tool_id,
                             'chat_id', r.chat_id,
                             'user_id', r.user_id,
                             'model_id', r.model_id,
                             'created_at', r.created_at,
                             'updated_at', r.updated_at,
                             'model', jsonb_strip_nulls(
                                     jsonb_build_object(
                                             'id', m.id,
                                             'name', m.name,
                                             'min_level', m.min_level,
                                             'created_at', m.created_at,
                                             'updated_at', m.updated_at
                                     )
                                      ),
                             'medias', coalesce(rm_agg.result_medias, '[]'::jsonb)
                     )
             ) as result
         from results r
                  left join models m on m.id = r.model_id
                  left join result_medias_agg rm_agg on rm_agg.result_id = r.id
     )
select
    c.id as chat_id,
    c.title as chat_title,
    c.created_at as chat_created_at,
    c.updated_at as chat_updated_at,
    c.user_id as chat_user_id,
    coalesce(
                    jsonb_agg(
                    jsonb_strip_nulls(
                            jsonb_build_object(
                                    'command', jsonb_build_object(
                                    'id', cm.id,
                                    'prompt', cm.prompt,
                                    'settings', cm.settings,
                                    'flags', cm.flags,
                                    'chat_id', cm.chat_id,
                                    'tool_id', cm.tool_id,
                                    'model_id', cm.model_id,
                                    'status', cm.status,
                                    'user_id', cm.user_id,
                                    'created_at', cm.created_at,
                                    'updated_at', cm.created_at,
                                    'medias', coalesce(cmm_agg.command_medias, '[]'::jsonb)
                                               ),
                                    'result', rpc.result
                            )
                    )
						ORDER BY cm.created_at, cm.id
                             ) filter (where cm.id is not null),
                    '[]'::jsonb
    ) as command_result_pairs
from chats c
         left join commands cm on cm.chat_id = c.id
         left join command_medias_agg cmm_agg on cmm_agg.command_id = cm.id
         left join result_per_command rpc on rpc.command_id = cm.id
where c.id = $1
  and c.user_id = $2
group by c.id, c.title, c.created_at, c.updated_at, c.user_id;`
