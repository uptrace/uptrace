package unixtime

import "time"

const (
    goFormat = "2006-01-02 15:04:05.999999999 -0700 MST"
    pythonFormat = "2006-01-02 15:04:05-07:00"
)

%% machine read_time;
%% write data;

func ReadTime(data string) (int, string) {
    cs, p, pe, eof := 0, 0, len(data), len(data)
    var ts, te, act int
    _, _, _ = eof, ts, act

%%{
    sp = ' ' ;
    int = digit+ ;
    float = digit+ '.' digit+ ;

    year4 = digit{4} ;
    month2 = digit{2} ;
    month_short = alpha{3} ;
    day2 = digit{2} ;
    yyyymmdd = year4 '-' month2 '-' day2 ;
    dayofweek_short = alpha{3} ;

    tz_name = alpha{2,5} ;
    tz_offset = [+\-]? digit{1,4} ;
    tz_hhmm = [+\-]? digit{2} ':' digit{2} ;

    date = dayofweek_short ' ' month_short ' '* digit{1,2} ;
    time = digit{2} ':' digit{2} ':' digit{2} ;

    ansic = date ' ' time ' ' year4 ;
    unix = date ' ' time ' ' tz_name ' ' year4 ;
    ruby = date ' ' time ' ' tz_offset ' ' year4 ;

    stamp = month_short ' '* digit{1,2} ' ' time ;
    nano3 = '.' digit{3} ;
    nano6 = '.' digit{6} ;
    nano9 = '.' digit{9} ;
    nano = '.' digit{1,9} ;

    rfc3339 = yyyymmdd 'T' time nano? ('Z' | tz_hhmm) ;
    golang = yyyymmdd sp time nano9 sp tz_offset sp tz_name ;
    golang_mono = golang sp 'm=' [+\-] float ;
    python = yyyymmdd sp time tz_hhmm ;
    postgres = yyyymmdd sp time nano3 ;
    android = month2 '-' day2 ' ' time ;
    nginx = year4 '/' month2 '/' day2 sp time ;

    main := |*
        ansic => { return te, time.ANSIC } ;
        unix => { return te, time.UnixDate } ;
        ruby => { return te, time.RubyDate } ;

        stamp => { return te, time.Stamp } ;
        stamp nano3 => { return te, time.StampMilli } ;
        stamp nano6 => { return te, time.StampMicro } ;
        stamp nano9 => { return te, time.StampNano } ;

        rfc3339 => { return te, time.RFC3339 } ;
        golang => { return te, goFormat } ;
        golang_mono => { return te, goFormat } ;
        python => { return te, pythonFormat } ;
        postgres => { return te, "2006-01-02 15:04:05.000" } ;
        postgres sp tz_name => { return te, "2006-01-02 15:04:05.000 MST" } ;
        android nano3 => { return te, "01-02 15:04:05.000" } ;
        nginx => { return te, "2006/01/02 15:04:05" } ;

        yyyymmdd => { return te, "2006-01-02" } ;
        month_short '-' day2 sp time nano3 => { return te, "Jan-02 15:04:05.000" } ;

        day2 '/' month_short '/' year4 ':' time => { return te, "02/Jan/2006:15:04:05" } ;
        day2 '/' month_short '/' year4 ':' time sp tz_offset => { return te, "02/Jan/2006:15:04:05 -0700" } ;
        day2 sp month_short sp year4 sp time nano3 => { return te, "02 Jan 2006 15:04:05.000" } ;
    *|;

    write init;
    write exec;
}%%

    return 0, ""
}
