import{C as S}from"./CodeBlock-B09wfSDq.js";import{d as z,r as t,o as s,m as d,w as n,b as i,e as l,l as R,y as V,T as E,k as m,t as v,c as _,F as N,s as A}from"./index-Be4yFAuI.js";const B=["data-testid","innerHTML"],I=z({__name:"ZoneConfigView",props:{data:{}},setup(f){const a=f;return(F,M)=>{const p=t("RouteTitle"),h=t("KAlert"),C=t("KCard"),k=t("AppView"),y=t("DataSource"),w=t("RouteView");return s(),d(w,{name:"zone-cp-config-view",params:{zone:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:o,t:c,uri:x})=>{var u,g;return[i(p,{render:!1,title:c("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"]),l(),i(y,{src:x(R(V),"/control-plane/outdated/:version",{version:((g=(u=a.data.zoneInsight.version)==null?void 0:u.kumaCp)==null?void 0:g.version)??"-"})},{default:n(({data:r})=>[i(k,null,E({title:n(()=>[m("h2",null,[i(p,{title:c("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"])])]),default:n(()=>[l(),l(),i(C,null,{default:n(()=>[Object.keys(a.data.zoneInsight.config).length>0?(s(),d(S,{key:0,language:"json",code:JSON.stringify(a.data.zoneInsight.config,null,2),"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:e=>o.update({codeSearch:e}),onFilterModeChange:e=>o.update({codeFilter:e}),onRegExpModeChange:e=>o.update({codeRegExp:e})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(s(),d(h,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{default:n(()=>[l(v(c("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))]),_:2},1024)]),_:2},[a.data.warnings.length>0?{name:"notifications",fn:n(()=>[m("ul",null,[(s(!0),_(N,null,A(a.data.warnings,e=>(s(),_("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:c(`common.warnings.${e.kind}`,{...e.payload,...e.kind==="INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS"?{globalCpVersion:(r==null?void 0:r.version)??""}:{}})},null,8,B))),128))])]),key:"0"}:void 0]),1024)]),_:2},1032,["src"])]}),_:1})}}});export{I as default};
