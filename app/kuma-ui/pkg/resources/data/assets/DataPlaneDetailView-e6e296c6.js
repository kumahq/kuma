import{d as N,r as ne,o as e,e as s,g as a,F as g,s as R,q as y,t as u,h as n,w as t,f as M,a as p,B as _e,b as d,Y as me,p as he,m as ve,$ as ce,j as q,c as F,J as ge,z as fe,u as ke,K as be,G as B,L as we,v as $e,M as Te}from"./index-483ef5d7.js";import{A as V,a as X,S as Oe,b as Pe}from"./SubscriptionHeader-46de8a77.js";import{f as J,m as ue,p as Y,E as Z,q as le,g as pe,e as De,D as j,S as Ee,o as se,A as Ae,_ as Le}from"./RouteView.vue_vue_type_script_setup_true_lang-4039cee4.js";import{_ as de}from"./CodeBlock.vue_vue_type_style_index_0_lang-742a791c.js";import{P as ye}from"./PolicyTypeTag-1586ecd2.js";import{T as H}from"./TagList-4e0afce4.js";import{t as oe,_ as Se}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-12c1a8bc.js";import{E as ae}from"./EnvoyData-448b1fff.js";import{T as Ce}from"./TabsWidget-c5984124.js";import{T as xe}from"./TextWithCopyButton-814e8ee9.js";import{_ as Re}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-7ce35e48.js";import{a as Ge,d as re,b as Ie,p as Me,c as Ne,C as qe,I as Be,e as je}from"./dataplane-30467516.js";import{_ as Fe}from"./RouteTitle.vue_vue_type_script_setup_true_lang-c3261a42.js";import"./CopyButton-7d479f6b.js";const U=v=>(he("data-v-d6898838"),v=v(),ve(),v),Ke={class:"mesh-gateway-policy-list"},ze=U(()=>y("h3",{class:"mb-2"},`
      Gateway policies
    `,-1)),He={key:0},Ue=U(()=>y("h3",{class:"mt-6 mb-2"},`
      Listeners
    `,-1)),We=U(()=>y("b",null,"Host",-1)),Ye=U(()=>y("h4",{class:"mt-2 mb-2"},`
              Routes
            `,-1)),Ze={class:"dataplane-policy-header"},Je=U(()=>y("b",null,"Route",-1)),Qe=U(()=>y("b",null,"Service",-1)),Ve={key:0,class:"badge-list"},Xe={class:"mt-1"},et=N({__name:"MeshGatewayDataplanePolicyList",props:{meshGatewayDataplane:{type:Object,required:!0},meshGatewayListenerEntries:{type:Array,required:!0},meshGatewayRoutePolicies:{type:Array,required:!0}},setup(v){const l=v;return(r,E)=>{const O=ne("router-link");return e(),s("div",Ke,[ze,a(),v.meshGatewayRoutePolicies.length>0?(e(),s("ul",He,[(e(!0),s(g,null,R(v.meshGatewayRoutePolicies,(f,w)=>(e(),s("li",{key:w},[y("span",null,u(f.type),1),a(`:

        `),n(O,{to:f.route},{default:t(()=>[a(u(f.name),1)]),_:2},1032,["to"])]))),128))])):M("",!0),a(),Ue,a(),y("div",null,[(e(!0),s(g,null,R(l.meshGatewayListenerEntries,(f,w)=>(e(),s("div",{key:w},[y("div",null,[y("div",null,[We,a(": "+u(f.hostName)+":"+u(f.port)+" ("+u(f.protocol)+`)
          `,1)]),a(),f.routeEntries.length>0?(e(),s(g,{key:0},[Ye,a(),n(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,R(f.routeEntries,(h,$)=>(e(),p(V,{key:$},_e({"accordion-header":t(()=>[y("div",Ze,[y("div",null,[y("div",null,[Je,a(": "),n(O,{to:h.route},{default:t(()=>[a(u(h.routeName),1)]),_:2},1032,["to"])]),a(),y("div",null,[Qe,a(": "+u(h.service),1)])]),a(),h.policies.length>0?(e(),s("div",Ve,[(e(!0),s(g,null,R(h.policies,(i,k)=>(e(),p(d(me),{key:`${w}-${k}`},{default:t(()=>[a(u(i.type),1)]),_:2},1024))),128))])):M("",!0)])]),_:2},[h.policies.length>0?{name:"accordion-content",fn:t(()=>[y("ul",Xe,[(e(!0),s(g,null,R(h.policies,(i,k)=>(e(),s("li",{key:`${w}-${k}`},[a(u(i.type)+`:

                      `,1),n(O,{to:i.route},{default:t(()=>[a(u(i.name),1)]),_:2},1032,["to"])]))),128))])]),key:"0"}:void 0]),1024))),128))]),_:2},1024)],64)):M("",!0)])]))),128))])])}}});const tt=J(et,[["__scopeId","data-v-d6898838"]]),at={class:"policy-type-heading"},st={class:"policy-list"},nt={key:0},lt=N({__name:"PolicyTypeEntryList",props:{id:{type:String,required:!1,default:"entry-list"},policyTypeEntries:{type:Object,required:!0}},setup(v){const l=v,r=[{label:"From",key:"sourceTags"},{label:"To",key:"destinationTags"},{label:"On",key:"name"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:O}){return{class:`cell-${O}`}}return(O,f)=>{const w=ne("router-link");return e(),p(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,R(l.policyTypeEntries,(h,$)=>(e(),p(V,{key:$},{"accordion-header":t(()=>[y("h3",at,[n(ye,{"policy-type":h.type},{default:t(()=>[a(u(h.type)+" ("+u(h.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",st,[n(d(ce),{class:"policy-type-table",fetcher:()=>({data:h.connections,total:h.connections.length}),headers:r,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{sourceTags:t(({rowValue:i})=>[i.length>0?(e(),p(H,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),destinationTags:t(({rowValue:i})=>[i.length>0?(e(),p(H,{key:0,class:"tag-list",tags:i},null,8,["tags"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),name:t(({rowValue:i})=>[i!==null?(e(),s(g,{key:0},[a(u(i),1)],64)):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",nt,[(e(!0),s(g,null,R(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},[n(w,{to:k.route},{default:t(()=>[a(u(k.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:k})=>[i!==null?(e(),p(de,{key:0,id:`${l.id}-${$}-${k}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const it=J(lt,[["__scopeId","data-v-71c85650"]]),ot={class:"policy-type-heading"},rt={class:"policy-list"},ct={key:1,class:"tag-list-wrapper"},ut={key:0},pt={key:1},dt={key:0},yt={key:0},_t=N({__name:"RuleEntryList",props:{id:{type:String,required:!1,default:"entry-list"},ruleEntries:{type:Object,required:!0}},setup(v){const l=v,r=[{label:"Type",key:"type"},{label:"Addresses",key:"addresses"},{label:"Conf",key:"config"},{label:"Origin policies",key:"origins"}];function E({headerKey:O}){return{class:`cell-${O}`}}return(O,f)=>{const w=ne("router-link");return e(),p(X,{"initially-open":[],"multiple-open":""},{default:t(()=>[(e(!0),s(g,null,R(l.ruleEntries,(h,$)=>(e(),p(V,{key:$},{"accordion-header":t(()=>[y("h3",ot,[n(ye,{"policy-type":h.type},{default:t(()=>[a(u(h.type)+" ("+u(h.connections.length)+`)
          `,1)]),_:2},1032,["policy-type"])])]),"accordion-content":t(()=>[y("div",rt,[n(d(ce),{class:"policy-type-table",fetcher:()=>({data:h.connections,total:h.connections.length}),headers:r,"cell-attrs":E,"disable-pagination":"","is-clickable":""},{type:t(({rowValue:i})=>[i.sourceTags.length===0&&i.destinationTags.length===0?(e(),s(g,{key:0},[a(`
                —
              `)],64)):(e(),s("div",ct,[i.sourceTags.length>0?(e(),s("div",ut,[a(`
                  From

                  `),n(H,{class:"tag-list",tags:i.sourceTags},null,8,["tags"])])):M("",!0),a(),i.destinationTags.length>0?(e(),s("div",pt,[a(`
                  To

                  `),n(H,{class:"tag-list",tags:i.destinationTags},null,8,["tags"])])):M("",!0)]))]),addresses:t(({rowValue:i})=>[i.length>0?(e(),s("ul",dt,[(e(!0),s(g,null,R(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},u(k),1))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),origins:t(({rowValue:i})=>[i.length>0?(e(),s("ul",yt,[(e(!0),s(g,null,R(i,(k,A)=>(e(),s("li",{key:`${$}-${A}`},[n(w,{to:k.route},{default:t(()=>[a(u(k.name),1)]),_:2},1032,["to"])]))),128))])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),config:t(({rowValue:i,rowKey:k})=>[i!==null?(e(),p(de,{key:0,id:`${l.id}-${$}-${k}-code-block`,code:i,language:"yaml","show-copy-button":!1},null,8,["id","code"])):(e(),s(g,{key:1},[a(`
                —
              `)],64))]),_:2},1032,["fetcher"])])]),_:2},1024))),128))]),_:1})}}});const mt=J(_t,[["__scopeId","data-v-74be3da4"]]),ht=y("h2",{class:"visually-hidden"},`
    Policies
  `,-1),vt={key:0,class:"mt-2"},gt=y("h2",{class:"mb-2"},`
      Rules
    `,-1),ft=N({__name:"SidecarDataplanePolicyList",props:{dppName:{type:String,required:!0},policyTypeEntries:{type:Object,required:!0},ruleEntries:{type:Array,required:!0}},setup(v){const l=v;return(r,E)=>(e(),s(g,null,[ht,a(),n(it,{id:"policies","policy-type-entries":l.policyTypeEntries,"data-testid":"policy-list"},null,8,["policy-type-entries"]),a(),v.ruleEntries.length>0?(e(),s("div",vt,[gt,a(),n(mt,{id:"rules","rule-entries":l.ruleEntries,"data-testid":"rule-list"},null,8,["rule-entries"])])):M("",!0)],64))}}),kt={key:2,class:"policies-list"},bt={key:3,class:"policies-list"},wt=N({__name:"DataplanePolicies",props:{dataplaneOverview:{type:Object,required:!0},policyTypes:{type:Array,required:!0}},setup(v){const l=v,r=ue(),E=q(null),O=q([]),f=q([]),w=q([]),h=q([]),$=q(!0),i=q(null),k=F(()=>l.policyTypes.reduce((o,_)=>Object.assign(o,{[_.name]:_}),{}));ge(()=>l.dataplaneOverview.name,function(){A()}),A();async function A(){var o,_;i.value=null,$.value=!0,O.value=[],f.value=[],w.value=[],h.value=[];try{if(((_=(o=l.dataplaneOverview.dataplane.networking.gateway)==null?void 0:o.type)==null?void 0:_.toUpperCase())==="BUILTIN")E.value=await r.getMeshGatewayDataplane({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name}),w.value=Q(E.value),h.value=W(E.value.policies);else{const{items:c}=await r.getSidecarDataplanePolicies({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name});O.value=ee(c??[]);const{items:b}=await r.getDataplaneRules({mesh:l.dataplaneOverview.mesh,name:l.dataplaneOverview.name});f.value=x(b??[])}}catch(m){m instanceof Error?i.value=m:console.error(m)}finally{$.value=!1}}function Q(o){const _=[],m=o.listeners??[];for(const c of m)for(const b of c.hosts)for(const D of b.routes){const L=[];for(const S of D.destinations){const T=W(S.policies),I={routeName:D.route,route:{name:"policy-detail-view",params:{mesh:o.gateway.mesh,policyPath:"meshgatewayroutes",policy:D.route}},service:S.tags["kuma.io/service"],policies:T};L.push(I)}_.push({protocol:c.protocol,port:c.port,hostName:b.hostName,routeEntries:L})}return _}function W(o){if(o===void 0)return[];const _=[];for(const m of Object.values(o)){const c=k.value[m.type];_.push({type:m.type,name:m.name,route:{name:"policy-detail-view",params:{mesh:m.mesh,policyPath:c.path,policy:m.name}}})}return _}function ee(o){const _=new Map;for(const c of o){const{type:b,service:D}=c,L=typeof D=="string"&&D!==""?[{label:"kuma.io/service",value:D}]:[],S=b==="inbound"||b==="outbound"?c.name:null;for(const[T,I]of Object.entries(c.matchedPolicies)){_.has(T)||_.set(T,{type:T,connections:[]});const K=_.get(T),z=k.value[T];for(const ie of I){const G=C(ie,z,c,L,S);K.connections.push(...G)}}}const m=Array.from(_.values());return m.sort((c,b)=>c.type.localeCompare(b.type)),m}function C(o,_,m,c,b){const D=o.conf&&Object.keys(o.conf).length>0?oe(o.conf):null,S=[{name:o.name,route:{name:"policy-detail-view",params:{mesh:o.mesh,policyPath:_.path,policy:o.name}}}],T=[];if(m.type==="inbound"&&Array.isArray(o.sources))for(const{match:I}of o.sources){const z={sourceTags:[{label:"kuma.io/service",value:I["kuma.io/service"]}],destinationTags:c,name:b,config:D,origins:S};T.push(z)}else{const K={sourceTags:[],destinationTags:c,name:b,config:D,origins:S};T.push(K)}return T}function x(o){const _=new Map;for(const c of o){_.has(c.policyType)||_.set(c.policyType,{type:c.policyType,connections:[]});const b=_.get(c.policyType),D=k.value[c.policyType],L=P(c,D);b.connections.push(...L)}const m=Array.from(_.values());return m.sort((c,b)=>c.type.localeCompare(b.type)),m}function P(o,_){const{type:m,service:c,subset:b,conf:D}=o,L=b?Object.entries(b):[];let S,T;m==="ClientSubset"?L.length>0?S=L.map(([G,te])=>({label:G,value:te})):S=[{label:"kuma.io/service",value:"*"}]:S=[],m==="DestinationSubset"?L.length>0?T=L.map(([G,te])=>({label:G,value:te})):typeof c=="string"&&c!==""?T=[{label:"kuma.io/service",value:c}]:T=[{label:"kuma.io/service",value:"*"}]:m==="ClientSubset"&&typeof c=="string"&&c!==""?T=[{label:"kuma.io/service",value:c}]:T=[];const I=o.addresses??[],K=D&&Object.keys(D).length>0?oe(D):null,z=[];for(const G of o.origins)z.push({name:G.name,route:{name:"policy-detail-view",params:{mesh:G.mesh,policyPath:_.path,policy:G.name}}});return[{type:{sourceTags:S,destinationTags:T},addresses:I,config:K,origins:z}]}return(o,_)=>$.value?(e(),p(Y,{key:0})):i.value!==null?(e(),p(Z,{key:1,error:i.value},null,8,["error"])):O.value.length>0?(e(),s("div",kt,[n(ft,{"dpp-name":l.dataplaneOverview.name,"policy-type-entries":O.value,"rule-entries":f.value},null,8,["dpp-name","policy-type-entries","rule-entries"])])):w.value.length>0&&E.value!==null?(e(),s("div",bt,[n(tt,{"mesh-gateway-dataplane":E.value,"mesh-gateway-listener-entries":w.value,"mesh-gateway-route-policies":h.value},null,8,["mesh-gateway-dataplane","mesh-gateway-listener-entries","mesh-gateway-route-policies"])])):(e(),p(le,{key:4}))}});const $t=J(wt,[["__scopeId","data-v-f57d0877"]]),Tt={key:3},Ot=N({__name:"StatusInfo",props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:[Error,null],required:!1,default:null}},setup(v){return(l,r)=>(e(),s("div",null,[v.isLoading?(e(),p(Y,{key:0})):v.hasError||v.error!==null?(e(),p(Z,{key:1,error:v.error},null,8,["error"])):v.isEmpty?(e(),p(le,{key:2})):(e(),s("div",Tt,[fe(l.$slots,"default")]))]))}}),Pt={class:"stack"},Dt={class:"columns",style:{"--columns":"4"}},Et={class:"status-with-reason"},At=["href"],Lt={class:"columns",style:{"--columns":"3"}},St=N({__name:"DataPlaneDetails",props:{dataplaneOverview:{type:Object,required:!0}},setup(v){const l=v,{t:r,formatIsoDate:E}=pe(),O=ue(),f=ke(),w=De(),h=[{hash:"#overview",title:r("data-planes.routes.item.tabs.overview")},{hash:"#insights",title:r("data-planes.routes.item.tabs.insights")},{hash:"#dpp-policies",title:r("data-planes.routes.item.tabs.policies")},{hash:"#xds-configuration",title:r("data-planes.routes.item.tabs.xds_configuration")},{hash:"#envoy-stats",title:r("data-planes.routes.item.tabs.stats")},{hash:"#envoy-clusters",title:r("data-planes.routes.item.tabs.clusters")}],$=F(()=>Ge(l.dataplaneOverview.dataplane,l.dataplaneOverview.dataplaneInsight)),i=F(()=>re(l.dataplaneOverview.dataplane)),k=F(()=>Ie(l.dataplaneOverview.dataplaneInsight)),A=F(()=>Me(l.dataplaneOverview,E)),Q=F(()=>{var x;const C=Array.from(((x=l.dataplaneOverview.dataplaneInsight)==null?void 0:x.subscriptions)??[]);return C.reverse(),C}),W=F(()=>{var _;const C=((_=l.dataplaneOverview.dataplaneInsight)==null?void 0:_.subscriptions)??[];if(C.length===0)return[];const x=C[C.length-1];if(!("version"in x)||!x.version)return[];const P=[],o=x.version;if(o.kumaDp&&o.envoy){const m=Ne(o);m.kind!==qe&&m.kind!==Be&&P.push(m)}return w.getters["config/getMulticlusterStatus"]&&re(l.dataplaneOverview.dataplane).find(b=>b.label===be)&&typeof o.kumaDp.kumaCpCompatible=="boolean"&&!o.kumaDp.kumaCpCompatible&&P.push({kind:je,payload:{kumaDp:o.kumaDp.version}}),P});async function ee(C){const{mesh:x,name:P}=l.dataplaneOverview;return await O.getDataplaneFromMesh({mesh:x,name:P},C)}return(C,x)=>(e(),p(Ce,{tabs:h},{overview:t(()=>[y("div",Pt,[W.value.length>0?(e(),p(Re,{key:0,warnings:W.value,"data-testid":"data-plane-warnings"},null,8,["warnings"])):M("",!0),a(),n(d(B),null,{body:t(()=>[y("div",Dt,[n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.status")),1)]),body:t(()=>[y("div",Et,[n(Ee,{status:$.value.status},null,8,["status"]),a(),$.value.reason.length>0?(e(),p(d(we),{key:0,label:$.value.reason.join(", "),class:"reason-tooltip"},{default:t(()=>[n(d($e),{icon:"info",size:"20","hide-title":""})]),_:1},8,["label"])):M("",!0)])]),_:1}),a(),n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.name")),1)]),body:t(()=>[n(xe,{text:l.dataplaneOverview.name},null,8,["text"])]),_:1}),a(),n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.tags")),1)]),body:t(()=>[i.value.length>0?(e(),p(H,{key:0,tags:i.value},null,8,["tags"])):(e(),s(g,{key:1},[a(u(d(r)("common.detail.none")),1)],64))]),_:1}),a(),n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.dependencies")),1)]),body:t(()=>[k.value!==null?(e(),p(H,{key:0,tags:k.value},null,8,["tags"])):(e(),s(g,{key:1},[a(u(d(r)("common.detail.none")),1)],64))]),_:1})])]),_:1}),a(),y("div",null,[y("h3",null,u(d(r)("data-planes.detail.mtls")),1),a(),A.value===null?(e(),p(d(Te),{key:0,class:"mt-4",appearance:"danger"},{alertMessage:t(()=>[a(u(d(r)("data-planes.detail.no_mtls"))+` —
              `,1),y("a",{href:d(r)("data-planes.href.docs.mutual-tls"),class:"external-link",target:"_blank"},u(d(r)("data-planes.detail.no_mtls_learn_more",{product:d(r)("common.product.name")})),9,At)]),_:1})):(e(),p(d(B),{key:1,class:"mt-4"},{body:t(()=>[y("div",Lt,[n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.certificateExpirationTime")),1)]),body:t(()=>[a(u(A.value.certificateExpirationTime),1)]),_:1}),a(),n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.lastCertificateRegeneration")),1)]),body:t(()=>[a(u(A.value.lastCertificateRegeneration),1)]),_:1}),a(),n(j,null,{title:t(()=>[a(u(d(r)("http.api.property.certificateRegenerations")),1)]),body:t(()=>[a(u(A.value.certificateRegenerations),1)]),_:1})])]),_:1}))]),a(),y("div",null,[n(se,{src:`/meshes/${d(f).params.mesh}/dataplanes/${d(f).params.dataPlane}`},{default:t(({data:P,error:o})=>[o?(e(),p(Z,{key:0,error:o},null,8,["error"])):P===void 0?(e(),p(Y,{key:1})):(e(),s(g,{key:2},[y("h3",null,u(d(r)("data-planes.detail.configuration")),1),a(),n(Se,{id:"code-block-data-plane",class:"mt-4",resource:P,"resource-fetcher":ee,"is-searchable":""},null,8,["resource"])],64))]),_:1},8,["src"])])])]),insights:t(()=>[n(d(B),null,{body:t(()=>[n(Ot,{"is-empty":Q.value.length===0},{default:t(()=>[n(X,{"initially-open":0},{default:t(()=>[(e(!0),s(g,null,R(Q.value,(P,o)=>(e(),p(V,{key:o},{"accordion-header":t(()=>[n(Oe,{subscription:P},null,8,["subscription"])]),"accordion-content":t(()=>[n(Pe,{subscription:P,"is-discovery-subscription":""},null,8,["subscription"])]),_:2},1024))),128))]),_:1})]),_:1},8,["is-empty"])]),_:1})]),"dpp-policies":t(()=>[n(d(B),null,{body:t(()=>[n(se,{src:"/*/policy-types"},{default:t(({data:P,error:o})=>[o?(e(),p(Z,{key:0,error:o},null,8,["error"])):P===void 0?(e(),p(Y,{key:1})):P.policies.length===0?(e(),p(le,{key:2})):(e(),p($t,{key:3,"dataplane-overview":v.dataplaneOverview,"policy-types":P.policies},null,8,["dataplane-overview","policy-types"]))]),_:1})]),_:1})]),"xds-configuration":t(()=>[n(d(B),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/xds`,"query-key":"envoy-data-xds-data-plane"},null,8,["src"])]),_:1})]),"envoy-stats":t(()=>[n(d(B),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/stats`,"query-key":"envoy-data-stats-data-plane"},null,8,["src"])]),_:1})]),"envoy-clusters":t(()=>[n(d(B),null,{body:t(()=>[n(ae,{src:`/meshes/${l.dataplaneOverview.mesh}/dataplanes/${l.dataplaneOverview.name}/data-path/clusters`,"query-key":"envoy-data-clusters-data-plane"},null,8,["src"])]),_:1})]),_:1}))}});const Ct=J(St,[["__scopeId","data-v-12eaa68d"]]),Wt=N({__name:"DataPlaneDetailView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(v){const l=v,{t:r}=pe();return(E,O)=>(e(),p(Le,{name:"data-plane-detail-view","data-testid":"data-plane-detail-view"},{default:t(({route:f})=>[n(Ae,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:f.params.mesh}},text:f.params.mesh},{to:{name:`${l.isGatewayView?"gateways":"data-planes"}-list-view`,params:{mesh:f.params.mesh}},text:d(r)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[y("h1",null,[n(Fe,{title:d(r)(`${l.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:f.params.dataPlane}),render:!0},null,8,["title"])])]),default:t(()=>[a(),n(se,{src:`/meshes/${f.params.mesh}/dataplane-overviews/${f.params.dataPlane}`},{default:t(({data:w,error:h})=>[h?(e(),p(Z,{key:0,error:h},null,8,["error"])):w===void 0?(e(),p(Y,{key:1})):(e(),p(Ct,{key:2,"dataplane-overview":w,"data-testid":"detail-view-details"},null,8,["dataplane-overview"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{Wt as default};
