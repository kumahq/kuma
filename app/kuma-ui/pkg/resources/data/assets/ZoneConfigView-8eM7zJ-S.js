import{C}from"./CodeBlock-MXAK4X7p.js";import{d as k,a as t,o as a,b as c,w as o,e as l,m as d,f as r,t as y,F as x,c as p,E as w,$ as R}from"./index-Eji0C-Q5.js";const V=["data-testid","innerHTML"],S=k({__name:"ZoneConfigView",props:{data:{},notifications:{default:()=>[]}},setup(g){const i=g;return(z,E)=>{const m=t("RouteTitle"),u=t("KAlert"),f=t("KCard"),_=t("AppView"),h=t("RouteView");return a(),c(h,{name:"zone-cp-config-view",params:{zone:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:n,t:s})=>[l(_,null,R({title:o(()=>[d("h2",null,[l(m,{title:s("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"])])]),default:o(()=>[r(),r(),l(f,null,{default:o(()=>[Object.keys(i.data.zoneInsight.config).length>0?(a(),c(C,{key:0,language:"json",code:JSON.stringify(i.data.zoneInsight.config,null,2),"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter,"is-reg-exp-mode":n.params.codeRegExp,onQueryChange:e=>n.update({codeSearch:e}),onFilterModeChange:e=>n.update({codeFilter:e}),onRegExpModeChange:e=>n.update({codeRegExp:e})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(a(),c(u,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:o(()=>[r(y(s("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))]),_:2},1024)]),_:2},[i.notifications.length>0?{name:"notifications",fn:o(()=>[d("ul",null,[(a(!0),p(x,null,w(i.notifications,e=>(a(),p("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:s(`common.warnings.${e.kind}`,e.payload)},null,8,V))),128))])]),key:"0"}:void 0]),1024)]),_:1})}}});export{S as default};
