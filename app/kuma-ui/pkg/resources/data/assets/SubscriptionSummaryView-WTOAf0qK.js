import{d as A,r as m,o as l,q as y,w as t,b as i,m as f,t as n,e,Q as $,L as b,R as u,c,K as g,s as h,p as _,at as B,a as D,a1 as E}from"./index-TQmXQMPR.js";const F={class:"stack-with-borders"},N={key:1,class:"mt-8 stack-with-borders"},L=A({__name:"SubscriptionSummaryView",props:{data:{},routeName:{}},setup(w){const C=w;return(I,s)=>{const S=m("XSelect"),k=m("XLayout"),v=m("XAlert"),x=m("XCodeBlock"),z=m("AppView"),V=m("DataCollection"),R=m("RouteView");return l(),y(R,{name:C.routeName,params:{subscription:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:"structured"}},{default:t(({route:d,t:r})=>[i(V,{items:C.data,predicate:a=>a.id===d.params.subscription},{item:t(({item:a})=>[i(z,null,{title:t(()=>[f("h2",null,n(a.zoneInstanceId??a.globalInstanceId??a.controlPlaneInstanceId),1)]),default:t(()=>[s[18]||(s[18]=e()),i(k,{type:"stack"},{default:t(()=>[f("header",null,[i(k,{type:"separated",size:"max"},{default:t(()=>[f("h3",null,n(r("subscriptions.routes.item.config")),1),s[0]||(s[0]=e()),f("div",null,[i(S,{label:r("subscriptions.routes.item.format"),selected:d.params.format,onChange:o=>{d.update({format:o})}},$({_:2},[b(["structured","yaml"],o=>({name:`${o}-option`,fn:t(()=>[e(n(r(`subscriptions.routes.item.formats.${o}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)]),s[17]||(s[17]=e()),d.params.format==="structured"?(l(),y(k,{key:0,type:"stack","data-testid":"structured-view"},{default:t(()=>[f("div",F,[i(u,{layout:"horizontal"},{title:t(()=>[e(n(r("http.api.property.version")),1)]),body:t(()=>{var o,p;return[(l(!0),c(g,null,b([(p=(o=a.version)==null?void 0:o.kumaCp)==null?void 0:p.version],X=>(l(),c(g,null,[e(n(X??"-"),1)],64))),256))]}),_:2},1024),s[7]||(s[7]=e()),i(u,{layout:"horizontal"},{title:t(()=>[e(n(r("http.api.property.connectTime")),1)]),body:t(()=>[e(n(r("common.formats.datetime",{value:Date.parse(a.connectTime??"")})),1)]),_:2},1024),s[8]||(s[8]=e()),a.disconnectTime?(l(),y(u,{key:0,layout:"horizontal"},{title:t(()=>[e(n(r("http.api.property.disconnectTime")),1)]),body:t(()=>[e(n(r("common.formats.datetime",{value:Date.parse(a.disconnectTime)})),1)]),_:2},1024)):h("",!0),s[9]||(s[9]=e()),i(u,{layout:"horizontal"},{title:t(()=>[e(n(r("subscriptions.routes.item.headers.responses")),1)]),body:t(()=>{var o;return[(l(!0),c(g,null,b([((o=a.status)==null?void 0:o.total)??{}],p=>(l(),c(g,null,[e(n(p.responsesSent)+"/"+n(p.responsesAcknowledged),1)],64))),256))]}),_:2},1024),s[10]||(s[10]=e()),(l(),c(g,null,b(["zoneInstanceId","globalInstanceId","controlPlaneInstanceId"],o=>(l(),c(g,{key:typeof o},[a[o]?(l(),y(u,{key:0,layout:"horizontal"},{title:t(()=>[e(n(r(`http.api.property.${o}`)),1)]),body:t(()=>[e(n(a[o]),1)]),_:2},1024)):h("",!0)],64))),64)),s[11]||(s[11]=e()),i(u,{layout:"horizontal"},{title:t(()=>[e(n(r("http.api.property.id")),1)]),body:t(()=>[e(n(a.id),1)]),_:2},1024)]),s[16]||(s[16]=e()),Object.keys(a.status.acknowledgements).length===0?(l(),y(v,{key:0,variant:"info"},{icon:t(()=>[i(_(B))]),default:t(()=>[e(" "+n(r("common.detail.subscriptions.no_stats",{id:a.id})),1)]),_:2},1024)):(l(),c("div",N,[f("div",null,[D(I.$slots,"default")]),s[14]||(s[14]=e()),i(u,{class:"mt-4",layout:"horizontal"},{title:t(()=>[f("strong",null,n(r("subscriptions.routes.item.headers.type")),1)]),body:t(()=>[e(n(r("subscriptions.routes.item.headers.stat")),1)]),_:2},1024),s[15]||(s[15]=e()),(l(!0),c(g,null,b(Object.entries(a.status.acknowledgements??{}),([o,p])=>(l(),y(u,{key:o,layout:"horizontal"},{title:t(()=>[e(n(r(`http.api.property.${o}`)),1)]),body:t(()=>[e(n(p.responsesSent)+"/"+n(p.responsesAcknowledged),1)]),_:2},1024))),128))]))]),_:2},1024)):(l(),y(x,{key:1,language:"yaml",code:_(E).stringify(a.$raw),"is-searchable":"",query:d.params.codeSearch,"is-filter-mode":d.params.codeFilter,"is-reg-exp-mode":d.params.codeRegExp,onQueryChange:o=>d.update({codeSearch:o}),onFilterModeChange:o=>d.update({codeFilter:o}),onRegExpModeChange:o=>d.update({codeRegExp:o})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:3},8,["name"])}}});export{L as default};
