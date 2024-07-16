import{d as w,r,o as s,m as o,w as e,b as l,k as m,X as c,e as t,c as i,F as f,s as h,t as d,p,q as C}from"./index-R1UcVqV9.js";const B={class:"stack"},K={class:"columns"},g=w({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(y){const u=y;return(n,D)=>{const k=r("KTruncate"),_=r("KBadge"),V=r("KCard"),v=r("AppView"),b=r("RouteView");return s(),o(b,{name:"mesh-external-service-detail-view",params:{}},{default:e(()=>[l(v,null,{default:e(()=>[m("div",B,[l(V,null,{default:e(()=>[m("div",K,[u.data.status.addresses.length>0?(s(),o(c,{key:0},{title:e(()=>[t(`
                Addresses
              `)]),body:e(()=>[l(k,null,{default:e(()=>[(s(!0),i(f,null,h(u.data.status.addresses,a=>(s(),i("span",{key:a.hostname},d(a.hostname),1))),128))]),_:1})]),_:1})):p("",!0),t(),n.data.spec.match?(s(),o(c,{key:1,class:"port"},{title:e(()=>[t(`
                Port
              `)]),body:e(()=>[(s(!0),i(f,null,h([n.data.spec.match],a=>(s(),o(_,{key:a.port,appearance:"info"},{default:e(()=>[t(d(a.port)+"/"+d(a.protocol),1)]),_:2},1024))),128))]),_:1})):p("",!0),t(),n.data.spec.match?(s(),o(c,{key:2,class:"tls"},{title:e(()=>[t(`
                TLS
              `)]),body:e(()=>[l(_,{appearance:"neutral"},{default:e(()=>{var a;return[t(d((a=n.data.spec.tls)!=null&&a.enabled?"Enabled":"Disabled"),1)]}),_:1})]),_:1})):p("",!0),t(),typeof n.data.status.vip<"u"?(s(),o(c,{key:3,class:"ip"},{title:e(()=>[t(`
                VIP
              `)]),body:e(()=>[t(d(n.data.status.vip.ip),1)]),_:1})):p("",!0)])]),_:1})])]),_:1})]),_:1})}}}),N=C(g,[["__scopeId","data-v-a4b82fef"]]);export{N as default};
