import{d as g,h as u,aN as b,aK as B,aO as N,r as p,o as i,i as S,w as t,j as n,a8 as T,n as e,p as r,a9 as m,H as l,V as x,l as c,F as _,I as h,m as z}from"./index-9e09c995.js";import{_ as D}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-78412b30.js";import"./AccordionList-39b30b49.js";const $=["data-testid","innerHTML"],A={"data-testid":"detail-view-details",class:"stack"},H={class:"columns"},K={key:0},M=g({__name:"DetailView",props:{data:{},notifications:{default:()=>[]}},setup(f){const s=f,v=u(()=>b(s.data)),k=u(()=>B(s.data)),V=u(()=>N(s.data));return(L,Z)=>{const y=p("KCard"),w=p("AppView"),C=p("RouteView");return i(),S(C,{name:"zone-cp-detail-view"},{default:t(({t:a})=>[n(w,null,T({default:t(()=>{var o;return[e(),r("div",A,[n(y,null,{body:t(()=>[r("div",H,[n(m,null,{title:t(()=>[e(l(a("http.api.property.status")),1)]),body:t(()=>[n(x,{status:k.value},null,8,["status"])]),_:2},1024),e(),n(m,null,{title:t(()=>[e(l(a("http.api.property.type")),1)]),body:t(()=>[e(l(a(`common.product.environment.${v.value||"unknown"}`)),1)]),_:2},1024),e(),n(m,null,{title:t(()=>[e(l(a("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(l(V.value||a("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),e(),(i(!0),c(_,null,h([((o=s.data.zoneInsight)==null?void 0:o.subscriptions)??[]],d=>(i(),c(_,{key:d},[d.length>0?(i(),c("div",K,[r("h2",null,l(a("zone-cps.detail.subscriptions")),1),e(),n(y,{class:"mt-4"},{body:t(()=>[n(D,{subscriptions:d},null,8,["subscriptions"])]),_:2},1024)])):z("",!0)],64))),128))])]}),_:2},[s.notifications.length>0?{name:"notifications",fn:t(()=>[r("ul",null,[(i(!0),c(_,null,h(s.notifications,o=>(i(),c("li",{key:o.kind,"data-testid":`warning-${o.kind}`,innerHTML:a(`common.warnings.${o.kind}`,o.payload)},null,8,$))),128)),e()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{M as default};
