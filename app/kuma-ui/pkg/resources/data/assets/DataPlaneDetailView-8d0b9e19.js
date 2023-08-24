import{f as J,m as ce,q as Y,E as Z,r as ne,g as ue,e as _e,D as q,S as me,K as he,o as se,A as ve,_ as ge}from"./RouteView.vue_vue_type_script_setup_true_lang-1dd6f4c1.js";import{d as M,r as le,o as e,e as s,g as a,F as g,s as x,q as y,t as u,h as n,w as t,f as N,a as p,B as fe,b as d,Y as ke,p as be,m as we,$ as pe,j as B,c as K,J as $e,z as Te,u as Oe,K as Pe,G as j,L as De,v as Ee,M as Ae}from"./index-30c3bdbc.js";import{A as V,a as X,S as Le,b as Se}from"./SubscriptionHeader-4cbb664c.js";import{_ as de}from"./CodeBlock.vue_vue_type_style_index_0_lang-bf980d9d.js";import{P as ye}from"./PolicyTypeTag-e709f9c7.js";import{T as U}from"./TagList-5cc25984.js";import{t as oe,_ as Ce}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3b5e22e9.js";import{E as ae}from"./EnvoyData-21b8c7a9.js";import{_ as Ie}from"./TabsWidget.vue_vue_type_style_index_0_lang-3e0839b4.js";import{T as xe}from"./TextWithCopyButton-19c2f2e4.js";import{_ as Re}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-bff247a4.js";import{a as Ge,d as re,b as Ne,p as Me,c as qe,C as Be,I as je,e as Ke}from"./dataplane-30467516.js";import{_ as Fe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cbf5001a.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b3f1f8ad.js";import"./CopyButton-bbd9c9e2.js";const H=v=>(be("data-v-23b43ae3"),v=v(),we(),v),ze={class:"mesh-gateway-policy-list"},Ue=H(()=>y("h3",{class:"mb-2"},`
      Gateway policies
    `,-1)),He={key:0},We=H(()=>y("h3",{class:"mt-6 mb-2"},`
      Listeners
    `,-1)),Ye=H(()=>y("b",null,"Host",-1)),Ze=H(()=>y("h4",{class:"mt-2 mb-2"},`
              Routes
            `,-1)),Je={class:"dataplane-policy-header"},Qe=H(()=>y("b",null,"Route",-1)),Ve=H(()=>y("b",null,"Service",-1)),Xe={key:0,class:"badge-list"},et={class:"mt-1"},tt=M({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(v){const l=v;return(r,E)=>{const O=le("router-link");return e(),s("div",ze,[Ue,a(),v.meshGatewayRoutePolicies.length>0?(e(),s("ul",He,[(e(!0),s(g,null,x(v.meshGatewayRoutePolicies,(f,w)=>(e(),s("li",{key:w},[y("span",null,u(f.type),1),a(`:

        `),n(O,{to:f.route},{default:t(()=>[a(u(f.name),1)]),_:2},1032,["to"])]))),128))])):N("",!0),a(),We,a(),y("div",null,[(e(!0),s(g,null,x(l.meshGatewayListenerEntries,(f,w)=>(e(),s("div",{key:w},[y("div",null,[y("div",null,[Ye,a(": "+u(f.hostName)+":"+u(f.port)+" ("+u(f.protocol)+`)
          `,1)]),a(),f.routeEntries.length>0?(e(),s(g,{key:0},[Ze,a(),n(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,x(f.routeEntries,(h,$)=>(e(),p(V,{key:$},fe({"accordion-header":t(()=>[y("div",Je,[y("div",null,[y("div",null,[Qe,a(": "),n(O,{to:h.route},{default:t(()=>[a(u(h.routeName),1)]),_:2},1032,["to"])]),a(),y("div",null,[Ve,a(": "+u(h.service),1)])]),a(),h.policies.length>0?(e(),s("div",Xe,[(e(!0),s(g,null,x(h.policies,(i,k)=>(e(),p(d(ke),{key:`${w}-${k}`},{default:t(()=>[a(u(i.type),1)]),_:2},1024))),128))])):N("",!0)])]),_:2},[h.policies.length>0?{name:"accordion-content",fn:t(()=>[y("ul",et,[(e(!0),s(g,null,x(h.policies,(i,k)=>(e(),s("li",{key:`${w}-${k}`},[a(u(i.type)+`:

                      `,1),n(O,{to:i.route},{default:t(()=>[a(u(i.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):N("",!0)])]))),128))])])}}});const at=J(tt,[["__scopeId","data-v-23b43ae3"]]),st={class:"policy-type-heading"},nt={class:"policy-list"},lt={key:0},it=M({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(v){const l=v,r=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:O}){return{class:`cell-${O}`}}return(O,f)=>{const w=le("router-link");return e(),p(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,x(l.policyTypeEntries,(h,$)=>(e(),p(V,{key:$},{"accordion-header":t(()=>[y("h3",st,[n(ye,{"policy-type":h.type},{default:t(()=>[a(u(h.type)+" ("+u(h.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",nt,[n(d(pe),{class:"policy-type-table",fetcher:()=>({data:h.connections,total:h.connections.length}),headers:r,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{sourceTags:t(({rowValue:i})=>[i.length>0?(e(),p(U,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),destinationTags:t(({rowValue:i})=>[i.length>0?(e(),p(U,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),name:t(({rowValue:i})=>[i!==null?(e(),s(g,{key:0},[a(u(i),1)],64)):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",lt,[(e(!0),s(g,null,x(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},[n(w,{to:k.route},{default:t(()=>[a(u(k.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:k})=>[i!==null?(e(),p(de,{key:0,id:`${l.id}-${$}-${k}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const ot=J(it,[["__scopeId","data-v-9a1971d5"]]),rt={class:"policy-type-heading"},ct={class:"policy-list"},ut={key:1,class:"tag-list-wrapper"},pt={key:0},dt={key:1},yt={key:0},_t={key:0},mt=M({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(v){const l=v,r=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:O}){return{class:`cell-${O}`}}return(O,f)=>{const w=le("router-link");return e(),p(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,x(l.ruleEntries,(h,$)=>(e(),p(V,{key:$},{"accordion-header":t(()=>[y("h3",rt,[n(ye,{"policy-type":h.type},{default:t(()=>[a(u(h.type)+" ("+u(h.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",ct,[n(d(pe),{class:"policy-type-table",fetcher:()=>({data:h.connections,total:h.connections.length}),headers:r,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{type:t(({rowValue:i})=>[i.sourceTags.length===0&&i.destinationTags.length===0?(e(),s(g,{key:0},[a(`
                —
              `)],64)):(e(),s("div",ut,[i.sourceTags.length>0?(e(),s("div",pt,[a(`
                  From

                  `),n(U,{class:"tag-list",tags:i.sourceTags},null,8,["tags"])])):N("",!0),a(),i.destinationTags.length>0?(e(),s("div",dt,[a(`
                  To

                  `),n(U,{class:"tag-list",tags:i.destinationTags},null,8,["tags"])])):N("",!0)]))]),addresses:t(({rowValue:i})=>[i.length>0?(e(),s("ul",yt,[(e(!0),s(g,null,x(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},u(k),1))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",_t,[(e(!0),s(g,null,x(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},[n(w,{to:k.route},{default:t(()=>[a(u(k.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:k})=>[i!==null?(e(),p(de,{key:0,id:`${l.id}-${$}-${k}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const ht=J(mt,[["__scopeId","data-v-3e59037c"]]),vt=y("h2",{class:"visually-hidden"},`
    Policies
  `,-1),gt={key:0,class:"mt-2"},ft=y("h2",{class:"mb-2"},`
      Rules
    `,-1),kt=M({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(v){const l=v;return(r,E)=>(e(),s(g,null,[vt,a(),n(ot,{id:"policies","policy-type-entries":l.policyTypeEntries,"data-testid":"policy-list"},null,8,["policy-type-entries"]),a(),v.ruleEntries.length>0?(e(),s("div",gt,[ft,a(),n(ht,{id:"rules","rule-entries":l.ruleEntries,"data-testid":"rule-list"},null,8,["rule-entries"])])):N("",!0)],64))}}),bt={key:2,class:"policies-list"},wt={key:3,class:"policies-list"},$t=M({__name:"DataplanePolicies",props:{dataplaneOverview:{type:Object,required:!0},policyTypes:{type:Array,required:!0}},setup(v){const l=v,r=ce(),E=B(null),O=B([]),f=B([]),w=B([]),h=B([]),$=B(!0),i=B(null),k=K(()=>l.policyTypes.reduce((o,_)=>Object.assign(o,{[_.name]:_}),{}));$e(()=>l.dataplaneOverview.name,function(){A()}),A();async function A(){var o,_;i.value=null,$.value=!0,O.value=[],f.value=[],w.value=[],h.value=[];try{if(((_=(o=l.dataplaneOverview.dataplane.networking.gateway)==null?void 0:o.type)==null?void 0:_.toUpperCase())==="BUILTIN")E.value=await r.getMeshGatewayDataplane({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name}),w.value=Q(E.value),h.value=W(E.value.policies);else{const{items:c}=await r.getSidecarDataplanePolicies({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name});O.value=ee(c??[]);const{items:b}=await r.getDataplaneRules({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name});f.value=I(b??[])}}catch(m){m instanceof Error?i.value=m:console.error(m)}finally{$.value=!1}}function Q(o){const _=[],m=o.listeners??[];for(const c of m)for(const b of c.hosts)for(const D of b.routes){const L=[];for(const S of D.destinations){const T=W(S.policies),G={routeName:D.route,route:{name:"policy-detail-view",params:{mesh:o.gateway.mesh,policyPath:"meshgatewayroutes",policy:D.route}},service:S.tags["kuma.io/service"],policies:T};L.push(G)}_.push({protocol:c.protocol,port:c.port,hostName:b.hostName,routeEntries:L})}return _}function W(o){if(o===void 0)return[];const _=[];for(const m of Object.values(o)){const c=k.value[m.type];_.push({type:m.type,name:m.name,route:{name:"policy-detail-view",params:{mesh:m.mesh,policyPath:c.path,policy:m.name}}})}return _}function ee(o){const _=new Map;for(const c of o){const{type:b,service:D}=c,L=typeof D=="string"&&D!==""?[{label:"kuma.io/service",value:D}]:[],S=b==="inbound"||b==="outbound"?c.name:null;for(const[T,G]of Object.entries(c.matchedPolicies)){_.has(T)||_.set(T,{type:T,connections:[]});const F=_.get(T),z=k.value[T];for(const ie of G){const R=C(ie,z,c,L,S);F.connections.push(...R)}}}const m=Array.from(_.values());return m.sort((c,b)=>c.type.localeCompare(b.type)),m}function C(o,_,m,c,b){const D=o.conf&&Object.keys(o.conf).length>0?oe(o.conf):null,S=[{name:o.name,route:{name:"policy-detail-view",params:{mesh:o.mesh,policyPath:_.path,policy:o.name}}}],T=[];if(m.type==="inbound"&&Array.isArray(o.sources))for(const{match:G}of o.sources){const z={sourceTags:[{label:"kuma.io/service",value:G["kuma.io/service"]}],destinationTags:c,name:b,config:D,origins:S};T.push(z)}else{const F={sourceTags:[],destinationTags:c,name:b,config:D,origins:S};T.push(F)}return T}function I(o){const _=new Map;for(const c of o){_.has(c.policyType)||_.set(c.policyType,{type:c.policyType,connections:[]});const b=_.get(c.policyType),D=k.value[c.policyType],L=P(c,D);b.connections.push(...L)}const m=Array.from(_.values());return m.sort((c,b)=>c.type.localeCompare(b.type)),m}function P(o,_){const{type:m,service:c,subset:b,conf:D}=o,L=b?Object.entries(b):[];let S,T;m==="ClientSubset"?L.length>0?S=L.map(([R,te])=>({label:R,value:te})):S=[{label:"kuma.io/service",value:"*"}]:S=[],m==="DestinationSubset"?L.length>0?T=L.map(([R,te])=>({label:R,value:te})):typeof c=="string"&&c!==""?T=[{label:"kuma.io/service",value:c}]:T=[{label:"kuma.io/service",value:"*"}]:m==="ClientSubset"&&typeof c=="string"&&c!==""?T=[{label:"kuma.io/service",value:c}]:T=[];const G=o.addresses??[],F=D&&Object.keys(D).length>0?oe(D):null,z=[];for(const R of o.origins)z.push({name:R.name,route:{name:"policy-detail-view",params:{mesh:R.mesh,policyPath:_.path,policy:R.name}}});return[{type:{sourceTags:S,destinationTags:T},addresses:G,config:F,origins:z}]}return(o,_)=>$.value?(e(),p(Y,{key:0})):i.value!==null?(e(),p(Z,{key:1,error:i.value},null,8,["error"])):O.value.length>0?(e(),s("div",bt,[n(kt,{"dpp-name":l.dataplaneOverview.name,"policy-type-entries":O.value,"rule-entries":f.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):w.value.length>0&&E.value!==null?(e(),s("div",wt,[n(at,{"mesh-gateway-dataplane":E.value,"mesh-gateway-listener-entries":w.value,"mesh-gateway-route-policies":h.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),p(ne,{key:4}))}});const Tt=J($t,[["__scopeId","data-v-5489bd97"]]),Ot={key:3},Pt=M({__name:"StatusInfo",props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:[Error,null],required:!1,default:null}},setup(v){return(l,r)=>(e(),s("div",null,[v.isLoading?(e(),p(Y,{key:0})):v.hasError||v.error!==null?(e(),p(Z,{key:1,error:v.error},null,8,["error"])):v.isEmpty?(e(),p(ne,{key:2})):(e(),s("div",Ot,[Te(l.$slots,"default")]))]))}}),Dt={class:"stack"},Et={class:"columns",style:{"--columns":"4"}},At={class:"status-with-reason"},Lt=["href"],St={class:"columns",style:{"--columns":"3"}},Ct=M({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(v){const l=v,{t:r,formatIsoDate:E}=ue(),O=ce(),f=Oe(),w=_e(),h=[{hash:"#overview",title:r("data-planes.routes.item.tabs.overview")},{hash:"#insights",title:r("data-planes.routes.item.tabs.insights")},{hash:"#dpp-policies",title:r("data-planes.routes.item.tabs.policies")},{hash:"#xds-configuration",title:r("data-planes.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:r("data-planes.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:r("data-planes.routes.item.tabs.clusters")}],$=K(()=>Ge(l.dataplaneOverview.dataplane,l.dataplaneOverview.dataplaneInsight)),i=K(()=>re(l.dataplaneOverview.dataplane)),k=K(()=>Ne(l.dataplaneOverview.dataplaneInsight)),A=K(()=>Me(l.dataplaneOverview,E)),Q=K(()=>{var I;const C=Array.from(((I=l.dataplaneOverview.dataplaneInsight)==null?void 0:I.subscriptions)??[]);return C.reverse(),C}),W=K(()=>{var _;const C=((_=l.dataplaneOverview.dataplaneInsight)==null?void 0:_.subscriptions)??[];if(C.length===0)return[];const I=C[C.length-1];if(!("version"in I)||!I.version)return[];const P=[],o=I.version;if(o.kumaDp&&o.envoy){const m=qe(o);m.kind!==Be&&m.kind!==je&&P.push(m)}return w.getters["config/getMulticlusterStatus"]&&re(l.dataplaneOverview.dataplane).find(b=>b.label===Pe)&&typeof o.kumaDp.kumaCpCompatible=="boolean"&&!o.kumaDp.kumaCpCompatible&&P.push({kind:Ke,payload:{kumaDp:o.kumaDp.version}}),P});async function ee(C){const{mesh:I,name:P}=l.dataplaneOverview;return await O.getDataplaneFromMesh({mesh:I,name:P},C)}return(C,I)=>(e(),p(Ie,{tabs:h},{overview:t(()=>[y("div",Dt,[W.value.length>0?(e(),p(Re,{key:0,warnings:W.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):N("",!0),a(),n(d(j),null,{body:t(()=>[y("div",Et,[n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.status")),1)]),body:t(()=>[y("div",At,[n(me,{status:$.value.status},null,8,["status"]),a(),$.value.reason.length>0?(e(),p(d(De),{key:0,label:$.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[n(d(Ee),{icon:"info",size:d(he),"hide-title":""},null,8,["size"])]),_:1},8,["label"])):N("",!0)])]),_:1}),a(),n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.name")),1)]),body:t(()=>[n(xe,{text:l.dataplaneOverview.name},null,8,["text"])]),_:1}),a(),n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.tags")),1)]),body:t(()=>[i.value.length>0?(e(),p(U,{key:0,tags:i.value},null,8,["tags"])):(e(),s(g,{key:1},[a(u(d(r)("common.detail.none")),1)],64))]),_:1}),a(),n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.dependencies")),1)]),body:t(()=>[k.value!==null?(e(),p(U,{key:0,tags:k.value},null,8,["tags"])):(e(),s(g,{key:1},[a(u(d(r)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),y("div",null,[y("h3",null,u(d(r)("data-planes.detail.mtls")),1),a(),A.value===null?(e(),p(d(Ae),{key:0,class:"mt-4",appearance:"danger"},{alertMessage:t(()=>[a(u(d(r)("data-planes.detail.no_mtls"))+` —
              `,1),y("a",{href:d(r)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},u(d(r)("data-planes.detail.no_mtls_learn_more",{product:d(r)("common.product.name")})),9,Lt)]),_:1})):(e(),p(d(j),{key:1,class:"mt-4"},{body:t(()=>[y("div",St,[n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.certificateExpirationTime")),1)]),body:t(()=>[a(u(A.value.certificateExpirationTime),1)]),_:1}),a(),n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.lastCertificateRegeneration")),1)]),body:t(()=>[a(u(A.value.lastCertificateRegeneration),1)]),_:1}),a(),n(q,null,{title:t(()=>[a(u(d(r)("http.api.property.certificateRegenerations")),1)]),body:t(()=>[a(u(A.value.certificateRegenerations),1)]),_:1})])]),_:1}))]),a(),y("div",null,[n(se,{src:`/meshes/${d(f).params.mesh}/dataplanes/${d(f).params.dataPlane}`},{default:t(({data:P,error:o})=>[o?(e(),p(Z,{key:0,error:o},null,8,["error"])):P===void 0?(e(),p(Y,{key:1})):(e(),s(g,{key:2},[y("h3",null,u(d(r)("data-planes.detail.configuration")),1),a(),n(Ce,{id:"code-block-data-plane",class:"mt-4",resource:P,"resource-fetcher":ee,"is-searchable":""},null,8,["resource"])],64))]),_:1},8,["src"])])])]),insights:t(()=>[n(d(j),null,{body:t(()=>[n(Pt,{"is-empty":Q.value.length===0},{default:t(()=>[n(X,{"initially-open":0},{default:t(()=>[(e(!0),s(g,null,x(Q.value,(P,o)=>(e(),p(V,{key:o},{"accordion-header":t(()=>[n(Le,{subscription:P},null,8,["subscription"])]),"accordion-content":t(()=>[n(Se,{subscription:P,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),_:1})]),"dpp-policies":t(()=>[n(d(j),null,{body:t(()=>[n(se,{src:"/*/policy-types"},{default:t(({data:P,error:o})=>[o?(e(),p(Z,{key:0,error:o},null,8,["error"])):P===void 0?(e(),p(Y,{key:1})):P.policies.length===0?(e(),p(ne,{key:2})):(e(),p(Tt,{key:3,"dataplane-overview":v.dataplaneOverview,"policy-types":P.policies},null,8,["dataplane-overview","policy-types"]))]),_:1})]),_:1})]),"xds-configuration":t(()=>[n(d(j),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/xds`,"query-key":"envoy-data-xds-data-plane"},null,8,["src"])]),_:1})]),"envoy-stats":t(()=>[n(d(j),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/stats`,"query-key":"envoy-data-stats-data-plane"},null,8,["src"])]),_:1})]),"envoy-clusters":t(()=>[n(d(j),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/clusters`,"query-key":"envoy-data-clusters-data-plane"},null,8,["src"])]),_:1})]),_:1}))}});const It=J(Ct,[["__scopeId","data-v-275fe936"]]),Zt=M({__name:"DataPlaneDetailView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(v){const l=v,{t:r}=ue();return(E,O)=>(e(),p(ge,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(({route:f})=>[n(ve,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:f.params.mesh}},text:f.params.mesh},{to:{name:`${l.isGatewayView?"gateways":"data-planes"}-list-view`,params:{mesh:f.params.mesh}},text:d(r)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[y("h1",null,[n(Fe,{title:d(r)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:f.params.dataPlane}),render:!0},null,8,["title"])])]),default:t(()=>[a(),n(se,{src:`/meshes/${f.params.mesh}/dataplane-overviews/${f.params.dataPlane}`},{default:t(({data:w,error:h})=>[h?(e(),p(Z,{key:0,error:h},null,8,["error"])):w===void 0?(e(),p(Y,{key:1})):(e(),p(It,{key:2,"dataplane-overview":w,"data-testid":"detail-view-details"},null,8,["dataplane-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{Zt as default};
