import{_ as S}from"./CodeBlock.vue_vue_type_style_index_0_lang-2bcf6524.js";import{d as V,a as i,o as n,b as g,w as o,e as m,a0 as z,p as h,f as l,c as d,F as f,B as y,t as F}from"./index-baa571c4.js";const M=["data-testid","innerHTML"],B=V({__name:"ZoneConfigView",props:{data:{},notifications:{default:()=>[]}},setup(C){const p=C;function k(_){var c;const s=((c=_.zoneInsight)==null?void 0:c.subscriptions)??[];if(s.length>0){const r=s[s.length-1];if(r.config)return JSON.stringify(JSON.parse(r.config),null,2)}return null}return(_,s)=>{const c=i("RouteTitle"),r=i("KAlert"),x=i("KCard"),w=i("AppView"),b=i("RouteView");return n(),g(b,{name:"zone-cp-config-view",params:{zone:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:t,t:u})=>[m(w,null,z({title:o(()=>[h("h2",null,[m(c,{title:u("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"])])]),default:o(()=>[l(),l(),m(x,null,{body:o(()=>[(n(!0),d(f,null,y([k(p.data)],(e,R)=>(n(),d(f,{key:R},[e!==null?(n(),g(S,{key:0,id:"code-block-zone-config",language:"json",code:e,"is-searchable":"",query:t.params.codeSearch,"is-filter-mode":t.params.codeFilter==="true","is-reg-exp-mode":t.params.codeRegExp==="true",onQueryChange:a=>t.update({codeSearch:a}),onFilterModeChange:a=>t.update({codeFilter:a}),onRegExpModeChange:a=>t.update({codeRegExp:a})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])):(n(),g(r,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:o(()=>[l(F(u("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))],64))),128))]),_:2},1024)]),_:2},[p.notifications.length>0?{name:"notifications",fn:o(()=>[h("ul",null,[(n(!0),d(f,null,y(p.notifications,e=>(n(),d("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:u(`common.warnings.${e.kind}`,e.payload)},null,8,M))),128)),l()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{B as default};
