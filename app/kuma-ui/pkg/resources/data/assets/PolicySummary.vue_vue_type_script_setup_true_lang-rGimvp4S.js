import{d as _,y as h,h as f,o,b as l,g as s,t,z as c,k as e,j as g,w as i,a as p,e as r,F as k,a1 as d,i as v}from"./index-9gITI0JG.js";const B={class:"stack"},C={key:0},R={class:"mt-4 stack"},b={key:0},N={class:"mt-4"},x=_({__name:"PolicySummary",props:{policy:{}},setup(m){const{t:n}=h(),a=m;return(y,V)=>{const u=f("KBadge");return o(),l("div",B,[a.policy.spec?(o(),l("div",C,[s("h3",null,t(c(n)("policies.routes.item.overview")),1),e(),s("div",R,[g(d,{layout:"horizontal"},{title:i(()=>[e(t(c(n)("http.api.property.targetRef")),1)]),body:i(()=>[a.policy.spec.targetRef?(o(),p(u,{key:0,appearance:"neutral"},{default:i(()=>[e(t(a.policy.spec.targetRef.kind),1),a.policy.spec.targetRef.name?(o(),l("span",b,[e(":"),s("b",null,t(a.policy.spec.targetRef.name),1)])):r("",!0)]),_:1})):(o(),l(k,{key:1},[e(t(c(n)("common.detail.none")),1)],64))]),_:1}),e(),a.policy.namespace.length>0?(o(),p(d,{key:0,layout:"horizontal"},{title:i(()=>[e(t(c(n)("data-planes.routes.item.namespace")),1)]),body:i(()=>[e(t(a.policy.namespace),1)]),_:1})):r("",!0)])])):r("",!0),e(),s("div",null,[s("h3",null,t(c(n)("policies.routes.item.config")),1),e(),s("div",N,[v(y.$slots,"default")])])])}}});export{x as _};
