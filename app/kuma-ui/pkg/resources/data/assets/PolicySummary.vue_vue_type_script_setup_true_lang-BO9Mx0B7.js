import{d as B,l as C,k as R,r as p,o as s,q as r,w as a,a as f,e as o,c as l,b as k,R as i,t as n,p as c,m as _,s as d}from"./index-TQmXQMPR.js";const X={key:0,class:"mt-4 stack-with-borders","data-testid":"structured-view"},w={key:0},N={key:1},V={class:"mt-4"},h=B({__name:"PolicySummary",props:{policy:{},format:{}},setup(g){const{t:y}=C(),z=R(),t=g;return(u,e)=>{const m=p("XBadge"),b=p("XAction"),v=p("XLayout");return s(),r(v,{type:"stack"},{default:a(()=>[f(u.$slots,"header"),e[8]||(e[8]=o()),t.policy.spec&&t.format==="structured"?(s(),l("div",X,[k(i,{layout:"horizontal"},{title:a(()=>[o(n(c(y)("http.api.property.targetRef")),1)]),body:a(()=>[t.policy.spec.targetRef?(s(),r(m,{key:0,appearance:"neutral"},{default:a(()=>[o(n(t.policy.spec.targetRef.kind),1),t.policy.spec.targetRef.name?(s(),l("span",w,[e[0]||(e[0]=o(":")),_("b",null,n(t.policy.spec.targetRef.name),1)])):d("",!0)]),_:1})):(s(),r(m,{key:1,appearance:"neutral"},{default:a(()=>e[1]||(e[1]=[o(`
              Mesh
            `)])),_:1}))]),_:1}),e[6]||(e[6]=o()),t.policy.namespace.length>0?(s(),r(i,{key:0,layout:"horizontal"},{title:a(()=>[o(n(c(y)("data-planes.routes.item.namespace")),1)]),body:a(()=>[o(n(t.policy.namespace),1)]),_:1})):d("",!0),e[7]||(e[7]=o()),c(z)("use zones")&&t.policy.zone?(s(),r(i,{key:1,layout:"horizontal"},{title:a(()=>e[4]||(e[4]=[o(`
            Zone
          `)])),body:a(()=>[k(b,{to:{name:"zone-cp-detail-view",params:{zone:t.policy.zone}}},{default:a(()=>[o(n(t.policy.zone),1)]),_:1},8,["to"])]),_:1})):d("",!0)])):(s(),l("div",N,[_("div",V,[f(u.$slots,"default")])]))]),_:3})}}});export{h as _};
