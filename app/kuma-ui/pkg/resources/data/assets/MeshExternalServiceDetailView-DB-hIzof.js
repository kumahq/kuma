import{d as g,r as c,o as a,m as n,w as e,b as i,k as m,Z as l,e as t,t as d,p,a2 as z,c as u,L as f,M as h,q as B}from"./index-BLKX6W_Q.js";const K={class:"stack"},A={class:"columns"},D=g({__name:"MeshExternalServiceDetailView",props:{data:{}},setup(y){const o=y;return(r,x)=>{const k=c("XAction"),v=c("KTruncate"),_=c("KBadge"),V=c("KCard"),w=c("AppView"),b=c("RouteView");return a(),n(b,{name:"mesh-external-service-detail-view"},{default:e(({can:C})=>[i(w,null,{default:e(()=>[m("div",K,[i(V,null,{default:e(()=>[m("div",A,[C("use zones")&&o.data.zone?(a(),n(l,{key:0},{title:e(()=>[t(`
                Zone
              `)]),body:e(()=>[i(k,{to:{name:"zone-cp-detail-view",params:{zone:o.data.zone}}},{default:e(()=>[t(d(o.data.zone),1)]),_:1},8,["to"])]),_:1})):p("",!0),t(),o.data.status.addresses.length>0?(a(),n(l,{key:1},{title:e(()=>[t(`
                Addresses
              `)]),body:e(()=>[o.data.status.addresses.length===1?(a(),n(z,{key:0,text:o.data.status.addresses[0].hostname},{default:e(()=>[t(d(o.data.status.addresses[0].hostname),1)]),_:1},8,["text"])):(a(),n(v,{key:1},{default:e(()=>[(a(!0),u(f,null,h(o.data.status.addresses,s=>(a(),u("span",{key:s.hostname},d(s.hostname),1))),128))]),_:1}))]),_:1})):p("",!0),t(),r.data.spec.match?(a(),n(l,{key:2,class:"port"},{title:e(()=>[t(`
                Port
              `)]),body:e(()=>[(a(!0),u(f,null,h([r.data.spec.match],s=>(a(),n(_,{key:s.port,appearance:"info"},{default:e(()=>[t(d(s.port)+"/"+d(s.protocol),1)]),_:2},1024))),128))]),_:1})):p("",!0),t(),r.data.spec.match?(a(),n(l,{key:3,class:"tls"},{title:e(()=>[t(`
                TLS
              `)]),body:e(()=>[i(_,{appearance:"neutral"},{default:e(()=>{var s;return[t(d((s=r.data.spec.tls)!=null&&s.enabled?"Enabled":"Disabled"),1)]}),_:1})]),_:1})):p("",!0),t(),typeof r.data.status.vip<"u"?(a(),n(l,{key:4,class:"ip"},{title:e(()=>[t(`
                VIP
              `)]),body:e(()=>[t(d(r.data.status.vip.ip),1)]),_:1})):p("",!0)])]),_:2},1024)])]),_:2},1024)]),_:1})}}}),N=B(D,[["__scopeId","data-v-f6c305e8"]]);export{N as default};
