package chproto

import (
	"time"
)

type ServerInfo struct {
	Name         string
	MinorVersion uint64
	MajorVersion uint64
	Revision     uint64
	Timezone     *time.Location
}

func (srv *ServerInfo) ReadFrom(rd *Reader) (err error) {
	if srv.Name, err = rd.String(); err != nil {
		return err
	}
	if srv.MajorVersion, err = rd.Uvarint(); err != nil {
		return err
	}
	if srv.MinorVersion, err = rd.Uvarint(); err != nil {
		return err
	}
	if srv.Revision, err = rd.Uvarint(); err != nil {
		return err
	}
	srv.Revision = min(srv.Revision, DBMS_TCP_PROTOCOL_VERSION)
	if srv.Revision >= DBMS_MIN_REVISION_WITH_SERVER_TIMEZONE {
		tz, err := rd.String()
		if err != nil {
			return err
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			return err
		}
		srv.Timezone = loc
	}
	if srv.Revision >= DBMS_MIN_REVISION_WITH_SERVER_DISPLAY_NAME {
		if _, err := rd.String(); err != nil {
			return err
		}
	}
	if srv.Revision >= DBMS_MIN_REVISION_WITH_VERSION_PATCH {
		if _, err := rd.Uvarint(); err != nil {
			return err
		}
	}
	return nil
}
