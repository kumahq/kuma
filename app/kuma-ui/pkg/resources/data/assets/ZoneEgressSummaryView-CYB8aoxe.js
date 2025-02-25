import{d as T,r as p,o as i,m as c,w as e,b as n,s as m,t as d,e as s,c as g,F as C,v as h,T as M,U as f,S as N,q as E,_ as Z}from"./index-Cm0u77zo.js";import{_ as $}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-M-Ig9rMl.js";const q={key:1},L={class:"mt-4"},Q=T({__name:"ZoneEgressSummaryView",props:{items:{}},setup(k){const z=k;return(I,o)=>{const S=p("XEmptyState"),w=p("RouteTitle"),b=p("XAction"),V=p("XSelect"),u=p("XLayout"),X=p("XCopyButton"),v=p("DataSource"),R=p("AppView"),B=p("DataCollection"),F=p("RouteView");return i(),c(F,{name:"zone-egress-summary-view",params:{proxy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,format:String}},{default:e(({route:a,t:r})=>[n(B,{items:z.items,predicate:y=>y.id===a.params.proxy,find:!0},{empty:e(()=>[n(S,null,{title:e(()=>[m("h2",null,d(r("common.collection.summary.empty_title",{type:"ZoneEgress"})),1)]),default:e(()=>[o[0]||(o[0]=s()),m("p",null,d(r("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),_:2},1024)]),default:e(({items:y})=>[(i(!0),g(C,null,h([y[0]],l=>(i(),c(R,{key:l.id},{title:e(()=>[m("h2",null,[n(b,{to:{name:"zone-egress-detail-view",params:{zone:l.zoneEgress.zone,proxyType:"egresses",proxy:l.id}}},{default:e(()=>[n(w,{title:r("zone-egresses.routes.item.title",{name:l.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[o[8]||(o[8]=s()),n(u,{type:"stack"},{default:e(()=>[m("header",null,[n(u,{type:"separated",size:"max"},{default:e(()=>[m("h3",null,d(r("zone-ingresses.routes.item.config")),1),o[1]||(o[1]=s()),m("div",null,[n(V,{label:r("zone-ingresses.routes.items.format"),selected:a.params.format,onChange:t=>{a.update({format:t})}},M({_:2},[h(["structured","yaml"],t=>({name:`${t}-option`,fn:e(()=>[s(d(r(`zone-ingresses.routes.items.formats.${t}`)),1)])}))]),1032,["label","selected","onChange"])])]),_:2},1024)]),o[7]||(o[7]=s()),a.params.format==="structured"?(i(),c(u,{key:0,type:"stack",class:"stack-with-borders","data-testid":"structured-view"},{default:e(()=>[n(f,{layout:"horizontal"},{title:e(()=>[s(d(r("http.api.property.status")),1)]),body:e(()=>[n(N,{status:l.state},null,8,["status"])]),_:2},1024),o[5]||(o[5]=s()),l.namespace.length>0?(i(),c(f,{key:0,layout:"horizontal"},{title:e(()=>[s(d(r("data-planes.routes.item.namespace")),1)]),body:e(()=>[s(d(l.namespace),1)]),_:2},1024)):E("",!0),o[6]||(o[6]=s()),n(f,{layout:"horizontal"},{title:e(()=>[s(d(r("http.api.property.address")),1)]),body:e(()=>[l.zoneEgress.socketAddress.length>0?(i(),c(X,{key:0,text:l.zoneEgress.socketAddress},null,8,["text"])):(i(),g(C,{key:1},[s(d(r("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1024)):(i(),g("div",q,[m("div",L,[n($,{resource:l.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:t=>a.update({codeSearch:t}),onFilterModeChange:t=>a.update({codeFilter:t}),onRegExpModeChange:t=>a.update({codeRegExp:t})},{default:e(({copy:t,copying:A})=>[A?(i(),c(v,{key:0,src:`/zone-egresses/${a.params.proxy}/as/kubernetes?no-store`,onChange:_=>{t(x=>x(_))},onError:_=>{t((x,D)=>D(_))}},null,8,["src","onChange","onError"])):E("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]))]),_:2},1024)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),G=Z(Q,[["__scopeId","data-v-db510d67"]]);export{G as default};
