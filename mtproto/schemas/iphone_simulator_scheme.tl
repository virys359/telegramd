///////////////////////////////
///////// Main application API
///////////////////////////////

---types---

schemeNotModified#263c9c58 = Scheme;
scheme#4e6ef65e scheme_raw:string types:Vector<SchemeType> methods:Vector<SchemeMethod> version:int = Scheme;
schemeParam#21b59bef name:string type:string = SchemeParam;
schemeMethod#479357c0 id:int method:string params:Vector<SchemeParam> type:string = SchemeMethod;
schemeType#a8e1e989 id:int predicate:string params:Vector<SchemeParam> type:string = SchemeType;

---functions---

help.getScheme#dbb69a9e version:int = Scheme;
