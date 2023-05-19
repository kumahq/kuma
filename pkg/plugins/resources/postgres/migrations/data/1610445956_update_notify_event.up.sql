CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$

DECLARE
    data json;
    notification json;

BEGIN

    -- Convert the old or new row to JSON, based on the kind of action.
    -- Action = DELETE?             -> OLD row
    -- Action = INSERT or UPDATE?   -> NEW row
    IF (TG_OP = 'DELETE') THEN
        data = json_build_object(
            'name', OLD.name,
            'mesh', OLD.mesh,
            'type', OLD.type);
        -- tenant_id is always empty so do not include it in JSON
    ELSE
        data = json_build_object(
                'name', NEW.name,
                'mesh', NEW.mesh,
                'type', NEW.type);
        -- tenant_id is always empty so do not include it in JSON
    END IF;

    -- Construct the notification as a JSON string.
    notification = json_build_object(
            'action', TG_OP,
            'data', data);


    -- Execute pg_notify(channel, notification)
    PERFORM pg_notify('resource_events',notification::text);

    -- Result is ignored since this is an AFTER trigger
    RETURN NULL;
END;

$$ LANGUAGE plpgsql;
